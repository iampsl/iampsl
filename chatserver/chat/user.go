package chat

import (
	"chatserver/global"
	"chatserver/mymsg"
	"chatserver/mysocket"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"unicode/utf8"

	"github.com/gorilla/websocket"

	"mvdan.cc/xurls"
)

//Start 启动
func Start() {
	initRedisMgr()
	initMysqlMgr()
	initRoomMgr()
	initUserMgr()
	initKeysMgr()
	log.Println("initialization success")
	go socketMgrTimeoutCheck()
	go socketListen()
	go httpListen()
	webSocketListen()
}

func socketListen() {
	listener, err := net.Listen("tcp", global.AppConfig.ListenIP+":"+strconv.Itoa(int(global.AppConfig.ListenPort)))
	if err != nil {
		log.Fatalln(err)
	}
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			global.AppLog.PrintlnError(err)
		} else {
			go onSocket(tcpConn)
		}
	}
}

func httpListen() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", onHTTPRequest)
	log.Fatalln(http.ListenAndServe(global.AppConfig.HTTPListenIP+":"+strconv.Itoa(int(global.AppConfig.HTTPListenPort)), mux))
}

func webSocketListen() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", onWebSocket)
	log.Fatalln(http.ListenAndServe(global.AppConfig.WsListenIP+":"+strconv.Itoa(int(global.AppConfig.WsListenPort)), mux))
}

var notifyMutex sync.Mutex

func onHTTPRequest(w http.ResponseWriter, r *http.Request) {
	type httpJSON struct {
		Result uint32 `json:"result"`
		Msg    string `json:"msg"`
	}
	if err := r.ParseForm(); err != nil {
		hj := httpJSON{Result: 1, Msg: "未知操作"}
		data, _ := json.Marshal(&hj)
		w.Write(data)
		return
	}
	cmd := r.Form["notify"]
	if len(cmd) != 1 {
		hj := httpJSON{Result: 1, Msg: "未知操作"}
		data, _ := json.Marshal(&hj)
		w.Write(data)
		return
	}
	if cmd[0] == "keywords" {
		notifyMutex.Lock()
		keys, err := getKeys()
		if err != nil {
			global.AppLog.PrintlnError(err)
			hj := httpJSON{Result: 2, Msg: "数据库失败"}
			data, _ := json.Marshal(&hj)
			w.Write(data)
		} else {
			keysmgr.SetKeys(keys)
			hj := httpJSON{Result: 0, Msg: "操作成功"}
			data, _ := json.Marshal(&hj)
			w.Write(data)
		}
		notifyMutex.Unlock()
	} else if cmd[0] == "users" {
		notifyMutex.Lock()
		limits, err := getLimits()
		if err != nil {
			global.AppLog.PrintlnError(err)
			hj := httpJSON{Result: 2, Msg: "数据库失败"}
			data, _ := json.Marshal(&hj)
			w.Write(data)
		} else {
			usrmgr.SetLimits(limits)
			hj := httpJSON{Result: 0, Msg: "操作成功"}
			data, _ := json.Marshal(&hj)
			w.Write(data)
		}
		notifyMutex.Unlock()
	} else {
		hj := httpJSON{Result: 1, Msg: "未知操作"}
		data, _ := json.Marshal(&hj)
		w.Write(data)
	}
}

func checkOrigin(r *http.Request) bool {
	return true
}

var upgrader = websocket.Upgrader{CheckOrigin: checkOrigin, Subprotocols: []string{"binary"}}

func onWebSocket(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	psocket := mysocket.NewMyWebSocket(c)
	socketmgr.AddSocket(psocket)
	ucontext := userContext{status: unlogin}
	ucontext.psocket = psocket
	defer clean(&ucontext)
	for {
		msgtype, msg, err := psocket.ReadMessage()
		if err != nil {
			global.AppLog.PrintlnInfo(err)
			break
		}
		if msgtype != websocket.BinaryMessage {
			global.AppLog.PrintlnInfo(msgtype, msg)
			break
		}
		var h mymsg.Head
		b, s := mymsg.UnSerializeHead(&h, msg)
		if !b {
			global.AppLog.PrintfInfo("UnSerializeHead failed\n")
			break
		}
		if int(h.Size) != len(msg) {
			global.AppLog.PrintfInfo("not equal\n")
			break
		}
		onMsg(h.Cmdid, msg[s:h.Size], psocket, &ucontext)
	}
}

type userStatus int

const (
	unlogin userStatus = iota
	login
)

type userContext struct {
	status         userStatus
	psocket        mysocket.MyWriteCloser
	puser          *UserInfo
	sitdown        bool
	roomID         uint16
	userInfoBuffer []*UserInfo
}

func clean(uc *userContext) {
	if uc.status == unlogin {
		socketmgr.RemoveSocket(uc.psocket)
	} else {
		usrmgr.RemoveUser(uc.puser)
		global.AppLog.PrintfInfo("clean %#v\n", uc.puser)
	}
	if uc.sitdown {
		roommgr.GetRoom(uc.roomID).RemoveUser(uc.puser)
	}
	uc.psocket.Close()
}

func onSocket(conn net.Conn) {
	psocket := mysocket.NewMySocket(conn)
	socketmgr.AddSocket(psocket)
	ucontext := userContext{status: unlogin}
	ucontext.psocket = psocket
	defer clean(&ucontext)
	const readBufferSize = 1024
	var readBuffer = make([]byte, readBufferSize)
	var readedSizes = 0
	for {
		if readedSizes == readBufferSize {
			global.AppLog.PrintfError("readBuffer reach limit\n")
			break
		}
		n, err := psocket.Read(readBuffer[readedSizes:])
		if err != nil {
			global.AppLog.PrintlnInfo(err)
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
	var h mymsg.Head
	b, s := mymsg.UnSerializeHead(&h, data)
	if !b {
		return 0
	}
	if int(h.Size) > readBufferSize {
		psocket.Close()
		global.AppLog.PrintfError("cmdid:%d cmdlen:%d buflen:%d\n", h.Cmdid, h.Size, readBufferSize)
		return 0
	}
	if int(h.Size) > len(data) {
		return 0
	}
	onMsg(h.Cmdid, data[s:h.Size], psocket, ucontext)
	return int(h.Size)
}

func onMsg(cmdid uint16, msg []byte, psocket mysocket.MyWriteCloser, ucontext *userContext) {
	if ucontext.status == unlogin {
		if cmdid != mymsg.CmdChatLogin {
			global.AppLog.PrintlnInfo(cmdid)
			psocket.Close()
		} else {
			onLogin(msg, psocket, ucontext)
		}
	} else {
		ucontext.puser.UpdateLastReadTime()
		switch cmdid {
		case mymsg.CmdChatSitDown:
			onSitDown(msg, psocket, ucontext)
		case mymsg.CmdChatSitUp:
			onSitUp(msg, psocket, ucontext)
		case mymsg.CmdChatSendMsgReq:
			onSendMsgReq(msg, psocket, ucontext)
		case mymsg.CmdChatHeartRsp:
		default:
			global.AppLog.PrintlnInfo(cmdid)
			psocket.Close()
		}
	}
}

func onLogin(msg []byte, psocket mysocket.MyWriteCloser, ucontext *userContext) {
	var loginMsg mymsg.ChatLogin
	if !loginMsg.UnSerialize(msg) {
		psocket.Close()
		global.AppLog.PrintfInfo("loginMsg.UnSerialize failed\n")
		return
	}
	global.AppLog.PrintfInfo("%#v\n", &loginMsg)
	userID, hallID, agentID, userName, hallName, agentName, passwd, state, rank, err := getUserInfo(loginMsg.Account, loginMsg.AgentCode)
	if err != nil {
		global.AppLog.PrintlnInfo(err)
		psocket.Close()
		return
	}
	if passwd != loginMsg.Password {
		global.AppLog.PrintfInfo("login:%v passwd:%s\n", loginMsg, passwd)
		psocket.Close()
		return
	}
	if state != 1 {
		global.AppLog.PrintlnInfo(state)
		psocket.Close()
		return
	}
	if b := socketmgr.RemoveSocket(psocket); !b {
		global.AppLog.PrintlnError(b)
		psocket.Close()
		return
	}
	ucontext.status = login
	ucontext.puser = NewUserInfo(psocket, loginMsg.Account, userName, userID, hallID, agentID, hallName, agentName, rank)
	old := usrmgr.AddUser(ucontext.puser)
	if old != nil {
		global.AppLog.PrintfInfo("again login\n")
		old.GetWriteCloser().Close()
	}
}

func onSitDown(msg []byte, psocket mysocket.MyWriteCloser, ucontext *userContext) {
	var sitDownMsg mymsg.ChatSitDown
	if !sitDownMsg.UnSerialize(msg) {
		global.AppLog.PrintfInfo("sitDownMsg.UnSerialize failed\n")
		psocket.Close()
		return
	}
	global.AppLog.PrintfInfo("%#v %#v %#v\n", &sitDownMsg, ucontext, ucontext.puser)
	room := roommgr.GetRoom(sitDownMsg.ServiceID)
	if room == nil {
		global.AppLog.PrintlnInfo(sitDownMsg.ServiceID)
		return
	}
	if ucontext.sitdown {
		if ucontext.roomID == sitDownMsg.ServiceID {
			return
		}
		roommgr.GetRoom(ucontext.roomID).RemoveUser(ucontext.puser)
		ucontext.sitdown = false
	}
	room.AddUser(ucontext.puser)
	ucontext.sitdown = true
	ucontext.roomID = sitDownMsg.ServiceID
}

func onSitUp(msg []byte, psocket mysocket.MyWriteCloser, ucontext *userContext) {
	var sitUpMsg mymsg.ChatSitUp
	if !sitUpMsg.UnSerialize(msg) {
		psocket.Close()
		global.AppLog.PrintfInfo("sitUpMsg.UnSerialize failed\n")
		return
	}
	global.AppLog.PrintfInfo("%#v %#v %#v\n", &sitUpMsg, ucontext, ucontext.puser)
	if !ucontext.sitdown {
		return
	}
	if ucontext.roomID != sitUpMsg.ServiceID {
		return
	}
	roommgr.GetRoom(ucontext.roomID).RemoveUser(ucontext.puser)
	ucontext.sitdown = false
}

func onSendMsgReq(msg []byte, psocket mysocket.MyWriteCloser, ucontext *userContext) {
	puser := ucontext.puser
	var sendMsgReq mymsg.ChatSendMsgReq
	if !sendMsgReq.UnSerialize(msg) {
		psocket.Close()
		global.AppLog.PrintfInfo("sendMsgReq.UnSerialize failed\n")
		return
	}
	global.AppLog.PrintfInfo("%#v %#v %#v\n", &sendMsgReq, ucontext, ucontext.puser)
	var sendMsgRsp mymsg.ChatSendMsgRsp
	if !ucontext.sitdown || ucontext.roomID != sendMsgReq.ServiceID {
		sendMsgRsp.Result = 1
		psocket.Write(&sendMsgRsp)
		return
	}
	if puser.GetRank() == 1 {
		sendMsgRsp.Result = 2
		psocket.Write(&sendMsgRsp)
		return
	}
	if usrmgr.isLimit(puser.GetHallID(), puser.GetUserID()) {
		sendMsgRsp.Result = 3
		psocket.Write(&sendMsgRsp)
		return
	}
	if valid := utf8.ValidString(sendMsgReq.Msg); !valid {
		sendMsgRsp.Result = 4
		psocket.Write(&sendMsgRsp)
		return
	}
	if isExistURL(sendMsgReq.Msg) {
		sendMsgRsp.Result = 5
		psocket.Write(&sendMsgRsp)
		notifyMutex.Lock()
		usrmgr.AddLimit(fmt.Sprintf("%d_%d", puser.GetHallID(), puser.GetUserID()))
		if err := addLimitUser(puser.GetUserID(), puser.GetUserName(), puser.GetHallID(), puser.GetAgentID(), puser.GetHallName(), puser.GetAgentName()); err != nil {
			global.AppLog.PrintlnError(err)
		}
		notifyMutex.Unlock()
		return
	}
	sendMsgRsp.Result = 0
	psocket.Write(&sendMsgRsp)

	var chatMsg mymsg.ChatMsg
	chatMsg.ServiceID = sendMsgReq.ServiceID
	chatMsg.Name = ucontext.puser.GetUserName()
	retMsg, b := keysmgr.Replace([]uint8(sendMsgReq.Msg), '*')
	if b {
		chatMsg.Msg = string(retMsg)
		err := addMonitor(puser.GetUserID(), puser.GetUserName(), puser.GetHallID(), puser.GetAgentID(), puser.GetHallName(), puser.GetAgentName(), sendMsgReq.Msg)
		if err != nil {
			global.AppLog.PrintlnError(err)
		}
	} else {
		chatMsg.Msg = sendMsgReq.Msg
	}
	ucontext.userInfoBuffer = roommgr.GetRoom(ucontext.roomID).GetAllUsers(ucontext.userInfoBuffer)
	for i := range ucontext.userInfoBuffer {
		ucontext.userInfoBuffer[i].GetWriteCloser().Write(&chatMsg)
		ucontext.userInfoBuffer[i] = nil
	}
}

var urlReg = xurls.Relaxed()

func isExistURL(msg string) bool {
	loc := urlReg.FindStringIndex(msg)
	if len(loc) > 0 {
		return true
	}
	return false
}
