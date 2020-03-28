
package main

import (
	"flag"
	"buffer"
	"time"

	"github.com/golang/glog"
)

func init() {
	glog.MaxSize = 1024 * 1024 * 100    //最大100M
	flag.Set("alsologtostderr", "true") // 日志写入文件的同时，输出到stderr
	flag.Set("log_dir", "./log")        // 日志文件保存目录
	flag.Set("v", "1")                  // 配置V输出的等级。
	flag.Parse()
}

func main() {
	pool, err := buffer.NewPool(10, 4096)
	if pool == nil {
		glog.Info(err)
		return
	}

	for i := 0; i < 1; i++ {
		go func() {
			glog.Info("Put Data begin..............")
			for i := 0; i < 51200; i++ {
				pool.Put(i)
			}
			glog.Info("Put Data End:", pool)
		}()
	}

	for i := 0; i < 1; i++ {
		go func() {
			glog.Info("Get Data begin..............")
			ticker := time.NewTicker(time.Second * 150)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					glog.Info("Get Data End:", pool)
					return
				default:
					pool.Get()
				}
			}
		}()
	}

	select {}
}
