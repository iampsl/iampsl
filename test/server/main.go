package main

import (
	"iampsl/public/mysocket"
	"log"
	"net"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	log.Println("server start")
	tcpAddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:9999")
	if err != nil {
		panic(err)
	}
	tcpListen, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		panic(err)
	}
	for {
		tcpConn, err := tcpListen.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		go tcpSession(tcpConn)
	}
}

func tcpSession(tcpConn *net.TCPConn) {
	psocket := mysocket.NewMySocket(tcpConn)
	defer psocket.Close()
	var readBuffer [1024]uint8
	readLen := 0
	for {
		n, err := psocket.Read(readBuffer[readLen:])
		if err != nil {
			break
		}
		readLen += n
		if readLen >= 10 {
			psocket.WriteBytes(readBuffer[0:readLen])
			readLen = 0
		}
	}
}
