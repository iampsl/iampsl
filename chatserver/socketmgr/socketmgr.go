package socketmgr

import (
	"chatServer/mysocket"
	"sync"
	"time"
)

type socketMgr struct {
	mutex   sync.Mutex
	sockets map[*mysocket.MySocket]int64
}

var smgr = &socketMgr{sockets: make(map[*mysocket.MySocket]int64)}

//AddSocket 添加一个socket
func AddSocket(psocket *mysocket.MySocket) {
	curTime := time.Now().Unix()
	smgr.mutex.Lock()
	smgr.sockets[psocket] = curTime
	smgr.mutex.Unlock()
}

//RemoveSocket 删除一个socket,返回值指示里面是否有这个socket
func RemoveSocket(psocket *mysocket.MySocket) bool {
	smgr.mutex.Lock()
	_, ok := smgr.sockets[psocket]
	if ok {
		delete(smgr.sockets, psocket)
	}
	smgr.mutex.Unlock()
	return ok
}

func removeTimeoutSockets(d time.Duration, out []*mysocket.MySocket) []*mysocket.MySocket {
	out = out[0:0]
	curTime := time.Now().Unix()
	smgr.mutex.Lock()
	for k, v := range smgr.sockets {
		if curTime-v > 10 {
			out = append(out, k)
		}
	}
	for _, v := range out {
		delete(smgr.sockets, v)
	}
	smgr.mutex.Unlock()
	return out
}

//TimeoutCheck 定时移除超时的socket
func TimeoutCheck() {
	tmpSockets := make([]*mysocket.MySocket, 0, 1000)
	for {
		time.Sleep(5 * time.Second)
		tmpSockets = removeTimeoutSockets(10*time.Second, tmpSockets)
		for i := range tmpSockets {
			tmpSockets[i].Close()
			tmpSockets[i] = nil
		}
	}
}
