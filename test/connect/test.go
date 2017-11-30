package connect

import (
	"fmt"
	"iampsl/public/mysocket"
	"iampsl/test/msg"
	"net"
)

func process(data []byte, psocket *mysocket.MySocket) int {
	msgHead := msg.Head{}
	b, headSize := msg.UnSerializeHead(&msgHead, data)
	if !b {
		return 0
	}
	if int(msgHead.Size) > len(data) {
		return 0
	}
	switch msgHead.Cmdid {
	case msg.CmdTryPlayRes:
		tryRes := msg.TryPlayRes{}
		if tryRes.UnSerialize(data[headSize:]) {
			fmt.Println(tryRes)
		} else {
			psocket.Close()
		}
	}
	return int(msgHead.Size)
}

//Test 测试入口
func Test(ch chan uint8) {
	conn, err := net.Dial("tcp", "192.168.31.230:8085")
	if err != nil {
		fmt.Println(err)
		return
	}
	psocket := mysocket.NewMySocket(conn)
	defer psocket.Close()
	var tryMsg msg.TryPlay
	tryMsg.LoginType = 3
	for i := 0; i < 1000; i++ {
		psocket.Write(&tryMsg)
	}
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
