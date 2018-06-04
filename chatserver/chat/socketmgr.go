package chat

import (
	"chatserver/mysocket"
	"sync"
	"time"
)

//SocketMgr 管理未登录的socket
type SocketMgr struct {
	mutex   sync.Mutex
	sockets map[mysocket.MyWriteCloser]int64
}

var socketmgr SocketMgr

func init() {
	socketmgr.sockets = make(map[mysocket.MyWriteCloser]int64)
}

//AddSocket 添加一个socket
func (mgr *SocketMgr) AddSocket(psocket mysocket.MyWriteCloser) {
	curTime := time.Now().Unix()
	mgr.mutex.Lock()
	mgr.sockets[psocket] = curTime
	mgr.mutex.Unlock()
}

//RemoveSocket 删除一个socket,返回值指示里面是否有这个socket
func (mgr *SocketMgr) RemoveSocket(psocket mysocket.MyWriteCloser) bool {
	mgr.mutex.Lock()
	_, ok := mgr.sockets[psocket]
	if ok {
		delete(mgr.sockets, psocket)
	}
	mgr.mutex.Unlock()
	return ok
}

//GetTimeoutSockets 得到超时sockets
func (mgr *SocketMgr) GetTimeoutSockets(d time.Duration, out []mysocket.MyWriteCloser) []mysocket.MyWriteCloser {
	out = out[0:0]
	curTime := time.Now().Unix()
	dseconds := int64(d.Seconds())
	mgr.mutex.Lock()
	for k, v := range mgr.sockets {
		if curTime-v > dseconds {
			out = append(out, k)
		}
	}
	mgr.mutex.Unlock()
	return out
}

func socketMgrTimeoutCheck() {
	tmpSockets := make([]mysocket.MyWriteCloser, 0, 1000)
	for {
		time.Sleep(5 * time.Second)
		tmpSockets = socketmgr.GetTimeoutSockets(10*time.Second, tmpSockets)
		for i := range tmpSockets {
			tmpSockets[i].Close()
			tmpSockets[i] = nil
		}
	}
}
