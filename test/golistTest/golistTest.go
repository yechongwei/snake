package main

import (
	"flag"
	"golist"
	"sync"
	"time"

	//"time"

	"github.com/golang/glog"
)

func init() {
	glog.MaxSize = 1024 * 104 * 100     //最大100M
	flag.Set("alsologtostderr", "true") // 日志写入文件的同时，输出到stderr
	flag.Set("log_dir", "./log")        // 日志文件保存目录
	flag.Set("v", "1")                  // 配置V输出的等级。
	flag.Parse()
}

func testlist1() {
	list := golist.NewList()

	//push
	glog.Info(list.RPush(1))
	glog.Info(list.RPush(2))
	glog.Info(list.RPush(3))
	glog.Info("rpush", list)
	glog.Info(list.LPush(4))
	glog.Info(list.LPush(5))
	glog.Info(list.LPush(6))
	glog.Info("lpush", list)

	glog.Info(list.Push(3, 7))
	glog.Info("push index3 7", list)
	glog.Info(list.Push(5, 8))
	glog.Info("push index3 8", list)
	glog.Info(list.Push(5, 9))
	glog.Info("push index3 9", list)

	//pop
	glog.Info(list.Pop(3))
	glog.Info(list.Pop(3))
	glog.Info(list.Pop(3))
	glog.Info("Pop(3)", list)
	glog.Info(list.LPop())
	glog.Info(list.LPop())
	glog.Info(list.LPop())
	glog.Info("lPop", list)
	glog.Info(list.RPop())
	glog.Info(list.RPop())
	glog.Info(list.RPop())
	glog.Info("RPop", list)
	glog.Info(list.Pop(0))
	glog.Info(list.Pop(1))
	glog.Info(list.Pop(2))
	glog.Info("Pop(3)", list)

}

func testlist2() {
	list := golist.NewList()

	glog.Info(list.Push(1, 1))
	glog.Info("push index1 1 ", list)
	glog.Info(list.Push(1, 2))
	glog.Info("push index1 2 ", list)
	glog.Info(list.Push(1, 3))
	glog.Info("push index1 3 ", list)
	glog.Info(list.Push(2, 4))
	glog.Info("push index2 5 ", list)
	glog.Info(list.Push(2, 5))
	glog.Info("push index1 3 ", list)
	glog.Info(list.Push(2, 6))
	glog.Info("push index2 6 ", list)
	glog.Info(list.Push(3, 7))
	glog.Info("push index3 7 ", list)
	glog.Info(list.Push(3, 8))
	glog.Info("push index3 8 ", list)
	glog.Info(list.Push(3, 9))
	glog.Info("push index3 9 ", list)

	glog.Info(list.Pop(4))
	glog.Info("Pop(4)", list)
	glog.Info(list.Pop(4))
	glog.Info("Pop(4)", list)
	glog.Info(list.Pop(4))
	glog.Info("Pop(4)", list)

	glog.Info(list.Pop(3))
	glog.Info("Pop(3)", list)
	glog.Info(list.Pop(2))
	glog.Info("Pop(2)", list)
	glog.Info(list.Pop(1))
	glog.Info("Pop(1)", list)
	glog.Info(list.Pop(1))
	glog.Info("Pop(1)", list)
	glog.Info(list.Pop(1))
	glog.Info("Pop(1)", list)
	glog.Info(list.Pop(1))
	glog.Info("Pop(1)", list)
	glog.Info(list.Pop(1))
	glog.Info("Pop(1)", list)

	glog.Info("3/2:", 3/2, "4/2", 4/2)
	var i int = 9
	glog.Info(i >> 1)
}

func testMatchlist3() {
	list := golist.NewList()

	glog.Info(list.Push(1, 1))
	glog.Info("push index1 1 ", list)
	glog.Info(list.Push(1, 1))
	glog.Info("push index1 1 ", list)
	glog.Info(list.Push(1, 3))
	glog.Info("push index1 3 ", list)

	fn := func(key, value interface{}) bool {
		value = value.(int)
		key = key.(int)
		return key == value
	}

	node := list.MatchAndRemove(1, fn)
	glog.Info("remove Node:", node, " isEmpty:", list.IsEmpty(), " | ", list)
	node = list.MatchAndRemove(1, fn)
	glog.Info("remove Node:", node, " isEmpty:", list.IsEmpty(), " | ", list)
	node = list.MatchAndRemove(3, fn)
	glog.Info("remove Node:", node, " isEmpty:", list.IsEmpty(), " | ", list)
}

func testlist4() {
	list := golist.NewList()
	var rwMutex sync.RWMutex
	for i := 0; i < 1; i++ {
		go func() {
			glog.Info("Put Data begin..............")
			for i := 0; i < 512000; i++ {
				rwMutex.Lock()
				list.Push(uint32(i%250000), i)
				rwMutex.Unlock()
			}
			glog.Info("Put Data End:", list)
		}()
	}

	for i := 0; i < 1; i++ {
		go func() {
			glog.Info("Get Data begin..............")
			ticker := time.NewTicker(time.Second * 60)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					glog.Info("Get Data End:", list)
					return
				default:
					rwMutex.Lock()
					list.Pop(0)
					rwMutex.Unlock()
				}
			}
		}()
	}
}

func main() {

	//testlist2()
	testMatchlist3()
	//testlist4()
	select {}
}
