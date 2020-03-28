package buffer

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// ErrClosedBuffer 是表示缓冲器已关闭的错误的变量。
var ErrClosedBuffer = errors.New("closed buffer")

//ErrBufferOverload 缓冲器已满 过载了
var ErrBufferOverload = errors.New("buffer is overload")

type IBuffer interface {
	Cap() uint32                        // Cap 用于获取本缓冲器的容量
	Len() uint32                        // Len 用于获取本缓冲器中的数据数量。
	Put(data interface{}) (bool, error) // Put 用于向缓冲器放入数据。
	Get() (interface{}, error)          // Get 用于从缓冲器获取器。
	Close() bool                        // Close 用于关闭缓冲器。 若缓冲器之前已关闭则返回false，否则返回true。
	Closed() bool                       // Closed 用于判断缓冲器是否已关闭。
}

// Buffer 代表缓冲器接口的实现类型。
type Buffer struct {
	// ch 代表存放数据的通道。
	ch chan interface{}
	// closed 代表缓冲器的关闭状态：0-未关闭；1-已关闭。
	closed uint32
	// closingLock 代表为了消除因关闭缓冲器而产生的竞态条件的读写锁。
	closingLock sync.RWMutex
}

// NewBuffer 用于创建一个缓冲器。参数size代表缓冲器的容量。
func NewBuffer(size uint32) (IBuffer, error) {
	if size == 0 {
		errMsg := fmt.Sprintf("illegal size for buffer: %d", size)
		return nil, errors.New(errMsg)
	}
	return &Buffer{
		ch: make(chan interface{}, size),
	}, nil
}

func (buf *Buffer) Cap() uint32 {
	return uint32(cap(buf.ch))
}

func (buf *Buffer) Len() uint32 {
	return uint32(len(buf.ch))
}

func (buf *Buffer) Put(data interface{}) (ok bool, err error) {
	//加锁处理  防止buf.ch被关闭时 出错
	buf.closingLock.RLock()
	defer buf.closingLock.RUnlock()
	if buf.Closed() {
		return false, ErrClosedBuffer
	}

	select {
	case buf.ch <- data:
		ok = true
	default:
		ok = false
		err = ErrBufferOverload
	}
	return
}

func (buf *Buffer) Get() (interface{}, error) {
	select {
	case data, ok := <-buf.ch:
		if !ok {
			return nil, ErrClosedBuffer
		}
		return data, nil
	default:
		return nil, errors.New("nothing")
	}
}

func (buf *Buffer) Close() bool {
	buf.closingLock.Lock()
	if atomic.CompareAndSwapUint32(&buf.closed, 0, 1) {
		close(buf.ch)
		buf.closingLock.Unlock()
		return true
	}
	buf.closingLock.Unlock()
	return false
}

func (buf *Buffer) Closed() bool {
	if atomic.LoadUint32(&buf.closed) == 0 {
		return false
	}
	return true
}
