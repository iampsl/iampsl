package msg

import (
	"encoding/binary"
	"iampsl/public/mybuffer"
)

func serializeHead(phead *head, pbuffer *mybuffer.MyBuffer) {
	pbuffer.AppendUint16(phead.size)
	pbuffer.AppendUint16(phead.cmdid)
}

//Head 消息头
type head struct {
	size  uint16
	cmdid uint16
}

// MSGHEADLEN 消息包头长度
const MSGHEADLEN int = 4

//TryPlay 试玩消息
type TryPlay struct {
	head
	LoginType uint8
}

//Serialize 系列化
func (pmsg *TryPlay) Serialize(pbuffer *mybuffer.MyBuffer) bool {
	pmsg.cmdid = 68
	begLen := pbuffer.Len()
	serializeHead(&(pmsg.head), pbuffer)
	pbuffer.AppendUint8(pmsg.LoginType)
	endLen := pbuffer.Len()
	data := pbuffer.Data()
	binary.LittleEndian.PutUint16(data[begLen:begLen+2], uint16(endLen-begLen))
	return true
}
