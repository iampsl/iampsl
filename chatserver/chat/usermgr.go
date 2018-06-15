package chat

import (
	"chatserver/mymsg"
	"chatserver/mysocket"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

//UserInfo 用户信息
type UserInfo struct {
	s            mysocket.MyWriteCloser
	lastReadTime int64
	account      string
	userName     string
	userID       uint32
	hallID       uint32
	agentID      uint32
	hallName     string
	agentName    string
	rank         int
}

//NewUserInfo 创建
func NewUserInfo(socket mysocket.MyWriteCloser, accountName string, uname string, uid uint32, hid uint32, agentid uint32, hname string, aname string, r int) *UserInfo {
	return &UserInfo{
		s:            socket,
		lastReadTime: time.Now().Unix(),
		account:      accountName,
		userName:     uname,
		userID:       uid,
		hallID:       hid,
		agentID:      agentid,
		hallName:     hname,
		agentName:    aname,
		rank:         r,
	}
}

//GetWriteCloser 得到socket
func (u *UserInfo) GetWriteCloser() mysocket.MyWriteCloser {
	return u.s
}

//GetAccount 得到帐号，带前缀
func (u *UserInfo) GetAccount() string {
	return u.account
}

//GetUserName 得到userName,不带前缀
func (u *UserInfo) GetUserName() string {
	return u.userName
}

//UpdateLastReadTime 更新最后读的时间
func (u *UserInfo) UpdateLastReadTime() {
	atomic.StoreInt64(&(u.lastReadTime), time.Now().Unix())
}

//GetLastReadTime 得到最后读的时间
func (u *UserInfo) GetLastReadTime() int64 {
	return atomic.LoadInt64(&(u.lastReadTime))
}

//GetUserID 得到userid
func (u *UserInfo) GetUserID() uint32 {
	return u.userID
}

//GetHallID 得到hallid
func (u *UserInfo) GetHallID() uint32 {
	return u.hallID
}

//GetAgentID 得到agentid
func (u *UserInfo) GetAgentID() uint32 {
	return u.agentID
}

//GetHallName 得到hallname
func (u *UserInfo) GetHallName() string {
	return u.hallName
}

//GetAgentName 得到agentname
func (u *UserInfo) GetAgentName() string {
	return u.agentName
}

//GetRank 得到rank
func (u *UserInfo) GetRank() int {
	return u.rank
}

//UserMgr 用户管理
type UserMgr struct {
	usersMutex sync.Mutex
	users      map[string]*UserInfo

	limitsMutex sync.RWMutex
	limits      map[string]uint8
}

var usrmgr UserMgr

func init() {
	usrmgr.users = make(map[string]*UserInfo)
	usrmgr.limits = make(map[string]uint8)
}

//AddUser 添加一个用户，返回值代表老的值
func (mgr *UserMgr) AddUser(u *UserInfo) *UserInfo {
	if u == nil {
		panic("u is nil")
	}
	key := fmt.Sprintf("%d_%d", u.GetHallID(), u.GetUserID())
	var limitMsg mymsg.ChatLimit
	if u.GetRank() == 1 {
		limitMsg.Limit = 1
	}
	mgr.usersMutex.Lock()
	old, _ := mgr.users[key]
	mgr.users[key] = u
	_, exist := mgr.limits[key]
	if exist {
		limitMsg.Limit = 1
	}
	u.GetWriteCloser().Write(&limitMsg)
	mgr.usersMutex.Unlock()
	return old
}

//RemoveUser 移除一个用户,返回值代表是否存在这个用户
func (mgr *UserMgr) RemoveUser(u *UserInfo) bool {
	if u == nil {
		panic("u is nil")
	}
	key := fmt.Sprintf("%d_%d", u.GetHallID(), u.GetUserID())
	ret := false
	mgr.usersMutex.Lock()
	v, b := mgr.users[key]
	if b && v == u {
		delete(mgr.users, key)
		ret = true
	}
	mgr.usersMutex.Unlock()
	return ret
}

func (mgr *UserMgr) heartBeat() {
	var heartReq mymsg.ChatHeartReq
	curTime := time.Now().Unix()
	mgr.usersMutex.Lock()
	for _, v := range mgr.users {
		sub := curTime - v.GetLastReadTime()
		if sub >= 60 {
			v.GetWriteCloser().Close()
		} else if sub >= 5 {
			v.GetWriteCloser().Write(&heartReq)
		}
	}
	mgr.usersMutex.Unlock()
}

//SetLimits 设置禁言用户
func (mgr *UserMgr) SetLimits(l map[string]uint8) {
	var limitMsg mymsg.ChatLimit
	mgr.usersMutex.Lock()
	mgr.limitsMutex.Lock()
	for k, v := range mgr.users {
		if v.GetRank() != 1 {
			_, old := mgr.limits[k]
			_, now := l[k]
			if old != now {
				if now {
					limitMsg.Limit = 1
				} else {
					limitMsg.Limit = 0
				}
				v.GetWriteCloser().Write(&limitMsg)
			}
		}
	}
	mgr.limits = l
	mgr.limitsMutex.Unlock()
	mgr.usersMutex.Unlock()
}

//AddLimit 添加禁言用户
func (mgr *UserMgr) AddLimit(u string) {
	var limitMsg mymsg.ChatLimit
	mgr.usersMutex.Lock()
	mgr.limitsMutex.Lock()
	v, ok := mgr.users[u]
	if !ok {
		mgr.limits[u] = 0
	} else {
		_, exist := mgr.limits[u]
		if !exist {
			mgr.limits[u] = 0
			if v.GetRank() != 1 {
				limitMsg.Limit = 1
				v.GetWriteCloser().Write(&limitMsg)
			}
		}
	}
	mgr.limitsMutex.Unlock()
	mgr.usersMutex.Unlock()
}

//isLimit 用户是否禁言
func (mgr *UserMgr) isLimit(hallid uint32, uid uint32) bool {
	s := fmt.Sprintf("%d_%d", hallid, uid)
	ret := false
	mgr.limitsMutex.RLock()
	_, ret = mgr.limits[s]
	mgr.limitsMutex.RUnlock()
	return ret
}

func initUserMgr() {
	limits, err := getLimits()
	if err != nil {
		log.Fatalln(err)
	}
	usrmgr.SetLimits(limits)
	go userMgrHeartBeat()
}

func userMgrHeartBeat() {
	for {
		time.Sleep(5 * time.Second)
		usrmgr.heartBeat()
	}
}
