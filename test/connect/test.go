package connect

import (
	"fmt"
	"iampsl/public/mysocket"
	"iampsl/test/msg"
	"net"
)

func process(data []byte, psocket *mysocket.MySocket) int {
	return len(data)
}

//Test 测试入口
func Test(ch chan int) {
	conn, err := net.Dial("tcp", "103.196.124.22:8085")
	if err != nil {
		fmt.Println(err)
		ch <- 0
		return
	}
	psocket := mysocket.NewMySocket(conn)
	defer psocket.Close()
	var tryMsg msg.TryPlay
	tryMsg.LoginType = 3
	psocket.Write(&tryMsg)
	var readBuffer = make([]byte, 10240)
	var readedSizes = 0
	for {
		n, err := psocket.Read(readBuffer[readedSizes:])
		if err != nil {
			fmt.Println(err)
			return
		}
		readedSizes += n
		procTotal := 0
		for {
			proc := process(readBuffer[procTotal:readedSizes], psocket)
			if proc == 0 {
				break
			}
			procTotal += proc
		}

		if procTotal > 0 {
			copy(readBuffer, readBuffer[procTotal:])
			readedSizes -= procTotal
		}
	}

}
