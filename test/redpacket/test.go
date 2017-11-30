package redpacket

import (
	"fmt"
	"iampsl/public/mysocket"
	"iampsl/test/msg"
	"net"
	"os"
	"time"
)

var rspNums int
var begTime time.Time

const totalNums = 100000

func process(data []byte, psocket *mysocket.MySocket) int {
	var msgHead msg.Head
	b, headSize := msg.UnSerializeHead(&msgHead, data)
	if !b {
		return 0
	}
	if int(msgHead.Size) > len(data) {
		return 0
	}
	switch msgHead.Cmdid {
	case msg.CmdServerRegisterRes:
		fmt.Println("服务器登录成功")
		clickMsg := msg.ClickRedPacket{RedID: 31, UID: 10, MoudleID: 8000, TotalBet: 10000.12}
		fmt.Printf("向服务器发送%v个抢红包请求\n", totalNums)
		begTime = time.Now()
		for i := 0; i < totalNums; i++ {
			if i%1000 == 0 {
				fmt.Printf("向服务器发送第%d个红包请求\n", i+1)
			}
			psocket.Write(&clickMsg)
		}
		fmt.Println("所有红包请求发送完成，等待红包响应...")
	case msg.CmdClickRedPacketRes:
		var clickResMsg msg.ClickRedPacketRes
		if clickResMsg.UnSerialize(data[headSize:]) {
			//fmt.Println(clickResMsg)
		} else {
			psocket.Close()
		}
		rspNums++
		if rspNums == totalNums {
			fmt.Printf("%v个红包响应完成,共用时%v秒\n", totalNums, time.Since(begTime).Seconds())
			os.Exit(0)
		}
	case msg.CmdKeepLive, 172:
	default:
		fmt.Printf("unknow cmd:%d\n", msgHead.Cmdid)
	}
	return int(msgHead.Size)
}

//Test 测试入口函数
func Test() {
	fmt.Println("连接服务器...")
	conn, err := net.Dial("tcp", "192.168.31.230:7777")
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("服务器连接成功")
	var readBuffer = make([]byte, 10240)
	var readedSizes = 0
	psocket := mysocket.NewMySocket(conn)
	registerMsg := msg.ServerRegister{MoudleID: 8000, MoudleType: 2, MoudlePort: 1000}
	fmt.Println("登录服务器")
	psocket.Write(&registerMsg)
	for {
		n, err := psocket.Read(readBuffer[readedSizes:])
		if err != nil {
			fmt.Println(err)
			psocket.Close()
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
