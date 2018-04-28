package mysocket

import (
	"chatServer/mybuffer"
	"net"
	"sync"
	"sync/atomic"
)

//MySocket 对net.Conn 的包装
type MySocket struct {
	conn      net.Conn
	buffers   [2]*mybuffer.MyBuffer
	sendIndex uint
	notify    chan int
	isclose   uint32

	m          sync.Mutex
	bclose     bool
	writeIndex uint
}

//NewMySocket 创建一个MySocket
func NewMySocket(c net.Conn) *MySocket {
	if c == nil {
		panic("c is nil")
	}
	var psocket = new(MySocket)
	psocket.conn = c
	psocket.buffers[0] = new(mybuffer.MyBuffer)
	psocket.buffers[1] = new(mybuffer.MyBuffer)
	psocket.sendIndex = 0
	psocket.notify = make(chan int, 1)
	psocket.isclose = 0
	psocket.bclose = false
	psocket.writeIndex = 1
	go doSend(psocket)
	return psocket
}

func doSend(my *MySocket) {
	writeErr := false
	for {
		_, ok := <-my.notify
		if !ok {
			return
		}
		my.m.Lock()
		my.writeIndex = my.sendIndex
		my.m.Unlock()
		my.sendIndex = (my.sendIndex + 1) % 2
		if !writeErr {
			var sendSplice = my.buffers[my.sendIndex].Data()
			for len(sendSplice) > 0 {
				n, err := my.conn.Write(sendSplice)
				if err != nil {
					writeErr = true
					break
				}
				sendSplice = sendSplice[n:]
			}
		}
		my.buffers[my.sendIndex].Clear()
	}
}

//Read 读数据
func (my *MySocket) Read(b []byte) (n int, err error) {
	return my.conn.Read(b)
}

//WriteBytes 写数据
func (my *MySocket) WriteBytes(b []byte) {
	if len(b) == 0 {
		return
	}
	my.m.Lock()
	if my.bclose {
		my.m.Unlock()
		return
	}
	var dataLen = my.buffers[my.writeIndex].Len()
	my.buffers[my.writeIndex].Append(b...)
	if dataLen == 0 {
		my.notify <- 0
	}
	my.m.Unlock()
}

//Serializer 系列化接口
type Serializer interface {
	Serialize(pbuffer *mybuffer.MyBuffer)
}

//Write 写数据
func (my *MySocket) Write(s Serializer) {
	if s == nil {
		return
	}
	my.m.Lock()
	if my.bclose {
		my.m.Unlock()
		return
	}
	var dataLen = my.buffers[my.writeIndex].Len()
	s.Serialize(my.buffers[my.writeIndex])
	var nowDataLen = my.buffers[my.writeIndex].Len()
	if dataLen == 0 && nowDataLen != 0 {
		my.notify <- 0
	}
	my.m.Unlock()
}

//Close 关闭一个MySocket, 释放系统资源
func (my *MySocket) Close() {
	my.m.Lock()
	if my.bclose {
		my.m.Unlock()
		return
	}
	my.bclose = true
	my.conn.Close()
	close(my.notify)
	my.m.Unlock()
	atomic.StoreUint32(&(my.isclose), 1)
}

//IsClose 判断MySocket是否关闭
func (my *MySocket) IsClose() bool {
	val := atomic.LoadUint32(&(my.isclose))
	if val > 0 {
		return true
	}
	return false
}
