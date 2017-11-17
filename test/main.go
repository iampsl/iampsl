// connect_svr 视讯接入服务器
// 作者：潘胜良

package main

import (
	"iampsl/test/connect"
)

const totalNums = 10000

var gchan = make(chan int, totalNums)

func main() {
	for i := 0; i < totalNums; i++ {
		gchan <- 0
	}
	for i := range gchan {
		i++
		go connect.Test(gchan)
	}
}
