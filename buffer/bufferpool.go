package buffer

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// ErrClosedBufferPool 是表示缓冲池已关闭的错误的变量。
var ErrClosedBufferPool = errors.New("pool is closed")

type IPool interface {
	//Cap 内存池中缓冲器的数量
	Cap() uint32
	//Len() 用于获取内存池中缓冲器的使用的数量
	Len() uint32
	// BufferCap 用于获取池中缓冲器的统一容量
	BufferCap() uint32
	// Total 用于获取缓冲池中数据的总数
	Total() uint64
	// Put 用于向缓冲池放入数据
	// 注意！本方法应该是阻塞的
	// 若缓冲池已关闭则会直接返回非nil的错误值
	Put(data interface{}) (ok bool, err error)
	// Get 用于从缓冲池获取数据
	// 注意！本方法应该是阻塞的
	// 若缓冲池已关闭则会直接返回非nil的错误值
	Get() (data interface{}, err error)
	// Close 用于关闭缓冲池。
	// 若缓冲池之前已关闭则返回false，否则返回true。
	Close() bool
	// Closed 用于判断缓冲池是否已关闭。
	Closed() bool
}

// BufferPool 代表数据缓冲池接口的实现类型。
type BufferPool struct {
	// poolCap 代表缓冲器的最大数量。
	poolCap uint32
	// poolSize 代表缓冲器的实际数量。
	poolSize uint32
	// bufferCap 代表缓冲器的统一容量。
	bufferCap uint32
	// total 代表池中数据的总数。
	total uint64
	// bufChs 代表存放缓冲器的通道。
	bufChs chan IBuffer
	// closed 代表缓冲池的关闭状态：0-未关闭；1-已关闭。
	closed uint32
	// lock 代表保护内部共享资源的读写锁。
	rwlock sync.RWMutex

	//putSize 存放了多少数据 用于测试
	putSize uint64
	//getSize 取了多少数据  用于测试
	getSize uint64
	//newBufferCount  创建buffer多少次  用于测试
	newBufferCount uint32
}

// NewPool 用于创建一个数据缓冲池
// 参数poolCap代表池中最多包含的缓冲器的数量
// 参数bufferCap代表池内缓冲器的统一容量
func NewPool(poolCap uint32, bufferCap uint32) (IPool, error) {
	if poolCap == 0 || bufferCap == 0 {
		errMsg := fmt.Sprintf("invalid params cannot eq 0 poolCap(%d) bufferCap(%d)", poolCap, bufferCap)
		return nil, errors.New(errMsg)
	}

	buffer, err := NewBuffer(bufferCap)
	if err != nil {
		return nil, err
	}

	bufChs := make(chan IBuffer, poolCap)
	bufChs <- buffer

	return &BufferPool{
		poolCap:   poolCap,
		poolSize:  1,
		bufferCap: bufferCap,
		total:     0,
		bufChs:    bufChs,
	}, nil
}

var fmtMsg = "cap(%d) len(%d) bufCap(%d) putSize(%d) getSize(%d) newBufCount(%d)"

func (pool *BufferPool) String() string {
	return fmt.Sprintf(fmtMsg, pool.Cap(), pool.Len(), pool.BufferCap(), pool.putSize, pool.getSize, pool.newBufferCount)
}

func (pool *BufferPool) BufferCap() uint32 {
	return pool.bufferCap
}

func (pool *BufferPool) Cap() uint32 {
	return pool.poolCap
}

func (pool *BufferPool) Len() uint32 {
	return atomic.LoadUint32(&pool.poolSize)
}

func (pool *BufferPool) Total() uint64 {
	return atomic.LoadUint64(&pool.total)
}

func (pool *BufferPool) Put(data interface{}) (ok bool, err error) {
	if pool.Closed() {
		return false, ErrClosedBufferPool
	}

	var count uint32
	var tryTimes uint32 = pool.Len()
	for buf := range pool.bufChs {
		ok, err = pool.putData(buf, data, &count, tryTimes)
		if err == nil || count > tryTimes {
			break
		}
	}
	return
}

// putData 用于向给定的缓冲器放入数据，并在必要时把缓冲器归还给池。
func (pool *BufferPool) putData(
	buf IBuffer, data interface{}, count *uint32, tryTimes uint32) (ok bool, err error) {
	if pool.Closed() {
		return false, ErrClosedBufferPool
	}

	defer func() {
		pool.rwlock.RLock()
		if pool.Closed() {
			buf.Close()
			atomic.AddUint32(&pool.poolSize, ^uint32(0))
		} else {
			pool.bufChs <- buf
		}
		pool.rwlock.RUnlock()
	}()

	if ok, err = buf.Put(data); ok {
		atomic.AddUint64(&pool.total, 1)
		atomic.AddUint64(&pool.putSize, 1)
		return
	}
	*count++

	if *count > tryTimes {
		pool.rwlock.Lock()
		if pool.Len() < pool.Cap() {
			if pool.Closed() {
				pool.rwlock.Unlock()
				return
			}
			newBuf, _ := NewBuffer(pool.bufferCap)
			ok, err = newBuf.Put(data)
			pool.bufChs <- newBuf
			atomic.AddUint32(&pool.poolSize, 1)
			atomic.AddUint64(&pool.total, 1)
			atomic.AddUint64(&pool.putSize, 1)
			atomic.AddUint32(&pool.newBufferCount, 1)
		}
		pool.rwlock.Unlock()
	}
	return
}

func (pool *BufferPool) Get() (data interface{}, err error) {
	if pool.Closed() {
		return false, ErrClosedBufferPool
	}

	var count uint32
	var tryTimes uint32 = pool.Len()
	for buf := range pool.bufChs {
		data, err = pool.getData(buf, &count, tryTimes)
		if err == nil || count > tryTimes {
			break
		}
	}
	return
}

// getData 用于从给定的缓冲器获取数据，并在必要时把缓冲器归还给池。
func (pool *BufferPool) getData(
	buf IBuffer, count *uint32, tryTimes uint32) (data interface{}, err error) {
	if pool.Closed() {
		buf.Close()
		return false, ErrClosedBufferPool
	}

	defer func() {
		//必须加上写锁 有可能pool.Len读取有问题 导致pool.bufChs一直为空 Get|Putt for buf := range pool.bufChs一直堵塞
		pool.rwlock.Lock()
		if pool.Closed() {
			buf.Close()
			atomic.AddUint32(&pool.poolSize, ^uint32(0))
			pool.rwlock.Unlock()
			return
		}

		if *count > tryTimes && buf.Len() == 0 && pool.Len() > 1 {
			buf.Close()
			//time.Sleep(time.Second * 10)
			//glog.Errorf("2 Count=%d, tryTimes=%d Closed", *count, tryTimes)
			atomic.AddUint32(&pool.poolSize, ^uint32(0))
		} else {
			pool.bufChs <- buf
		}
		pool.rwlock.Unlock()
	}()

	data, err = buf.Get()
	if err == nil {
		atomic.AddUint64(&pool.total, ^uint64(0))
		atomic.AddUint64(&pool.getSize, 1)
		return
	}
	*count++
	return
}

func (pool *BufferPool) Close() bool {
	pool.rwlock.Lock()
	if !atomic.CompareAndSwapUint32(&pool.closed, 0, 1) {
		pool.rwlock.Unlock()
		return false
	}
	close(pool.bufChs)
	pool.closeBufChans()
	pool.rwlock.Unlock()
	return true
}

func (pool *BufferPool) closeBufChans() {
	for buf := range pool.bufChs {
		buf.Close()
	}
}

//Closed  0-未关闭；1-已关闭
func (pool *BufferPool) Closed() bool {
	if atomic.LoadUint32(&pool.closed) == 0 {
		return false
	}
	return true
}
