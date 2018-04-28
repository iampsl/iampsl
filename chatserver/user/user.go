package user

import (
	"chatServer/global"
	"chatServer/mysocket"
	"chatServer/socketmgr"
	"log"
	"net"
	"strconv"
)

//Listen 侦听
func Listen() {
	listener, err := net.Listen("tcp", global.AppConfig.ListenIP+":"+strconv.Itoa(int(global.AppConfig.ListenPort)))
	if err != nil {
		log.Fatalln(err)
	}
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			global.AppLog.PrintlnError(err)
		} else {
			go onNewUser(tcpConn)
		}
	}

}

type userStatus int

const (
	unlogin userStatus = iota
	login
)

type userContext struct {
	status userStatus
}

func onNewUser(conn net.Conn) {
	psocket := mysocket.NewMySocket(conn)
	socketmgr.AddSocket(psocket)
	ucontext := userContext{status: unlogin}
	defer func() {
		if ucontext.status == unlogin {
			socketmgr.RemoveSocket(psocket)
		}
		psocket.Close()
	}()
	const readBufferSize = 10240
	var readBuffer = make([]byte, readBufferSize)
	var readedSizes = 0
	for {
		if readedSizes == readBufferSize {
			global.AppLog.PrintfError("readBuffer reach limit\n")
			break
		}
		n, err := psocket.Read(readBuffer[readedSizes:])
		if err != nil {
			break
		}
		readedSizes += n
		procTotal := 0
		for {
			if psocket.IsClose() {
				procTotal = readedSizes
				break
			}
			proc := process(readBuffer[procTotal:readedSizes], psocket, &ucontext, readBufferSize)
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

func process(data []byte, psocket *mysocket.MySocket, ucontext *userContext, readBufferSize int) int {
	return len(data)
}
