package main

import (
	"iampsl/public/mysocket"
	"log"
	"net"
	"sync"
)

var wg sync.WaitGroup

func connect() {
	defer wg.Done()
	conn, err := net.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		log.Println(err)
		return
	}
	psocket := mysocket.NewMySocket(conn)
	defer psocket.Close()
	var data [10]byte
	var readBuffer [1024]byte
	var readedSizes = 0
	psocket.WriteBytes(data[:])
	for {
		n, err := psocket.Read(readBuffer[readedSizes:])
		if err != nil {
			log.Println(err)
			return
		}
		readedSizes += n
		if readedSizes >= 10 {
			psocket.WriteBytes(readBuffer[0:readedSizes])
			readedSizes = 0
		}
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go connect()
	}
	wg.Wait()
}
