package chat

import (
	"log"
	"sync"
)

//RoomMgr 管理房间
type RoomMgr struct {
	rooms map[uint16]*Room
}

var roommgr = RoomMgr{rooms: make(map[uint16]*Room)}

//GetRoom 得到指定房间
func (mgr *RoomMgr) GetRoom(id uint16) *Room {
	return mgr.rooms[id]
}

func initRoomMgr() {
	ids, err := getAllRoomIDs()
	if err != nil {
		log.Fatalln(err)
	}
	for _, v := range ids {
		var r Room
		r.users = make(map[*UserInfo]uint8)
		roommgr.rooms[v] = &r
	}
}

//Room 房间
type Room struct {
	mu    sync.Mutex
	users map[*UserInfo]uint8
}

//AddUser 添加用户
func (r *Room) AddUser(u *UserInfo) {
	r.mu.Lock()
	r.users[u] = 0
	r.mu.Unlock()
}

//RemoveUser 移除用户
func (r *Room) RemoveUser(u *UserInfo) {
	r.mu.Lock()
	delete(r.users, u)
	r.mu.Unlock()
}

//GetAllUsers 得到所有用户
func (r *Room) GetAllUsers(out []*UserInfo) []*UserInfo {
	out = out[0:0]
	r.mu.Lock()
	for k := range r.users {
		out = append(out, k)
	}
	r.mu.Unlock()
	return out
}
