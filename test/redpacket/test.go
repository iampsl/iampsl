package redpacket

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"iampsl/public/mysocket"
	"net"
	"time"
)

type login struct {
	size uint16
	cmd  uint16
	mid  uint16
	mt   uint16
	port uint16
}

type clickRedPacket struct {
	size     uint16
	cmd      uint16
	redid    uint32
	uid      uint32
	serverid uint32
	totalbet float64
}

var rspNums int
var begTime time.Time

const totalNums = 100000

func process(data []byte, psocket *mysocket.MySocket) int {
	if len(data) < 4 {
		return 0
	}
	size := binary.LittleEndian.Uint16(data[:2])
	if int(size) > len(data) {
		return 0
	}
	cmd := binary.LittleEndian.Uint16(data[2:4])
	switch cmd {
	case 108:
		clickMsg := clickRedPacket{
			size:     24,
			cmd:      142,
			redid:    31,
			uid:      10,
			serverid: 8000,
			totalbet: 1000.0,
		}
		pbuffer := new(bytes.Buffer)
		err := binary.Write(pbuffer, binary.LittleEndian, clickMsg)
		if err != nil {
			fmt.Println(err)
		} else {
			begTime = time.Now()
			for i := 0; i < totalNums; i++ {
				psocket.Write(pbuffer.Bytes())
			}
		}
	case 143:
		rspNums++
		if rspNums == totalNums {
			fmt.Println(time.Since(begTime).Seconds())
		}
	case 11:
	default:
		fmt.Printf("unknow cmd:%d\n", cmd)
	}
	return int(size)
}

//Test 测试入口函数
func Test() {
	conn, err := net.Dial("tcp", "103.196.124.22:7777")
	if err != nil {
		fmt.Print(err)
		return
	}
	var readBuffer = make([]byte, 10240)
	var readedSizes = 0
	psocket := mysocket.NewMySocket(conn)
	var lg = login{size: 10, cmd: 107, mid: 8000, mt: 2, port: 1000}
	var pbuffer = new(bytes.Buffer)
	err = binary.Write(pbuffer, binary.LittleEndian, lg)
	if err != nil {
		fmt.Println(err)
	}
	psocket.Write(pbuffer.Bytes())
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
