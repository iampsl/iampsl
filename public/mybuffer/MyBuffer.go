package mybuffer

import (
	"encoding/binary"
)

// MyBuffer 我的缓冲
type MyBuffer struct {
	splice []byte
}

// Data 取得底层数据切片
func (pbuffer *MyBuffer) Data() []byte {
	return pbuffer.splice
}

//Clear 清空数据
func (pbuffer *MyBuffer) Clear() {
	pbuffer.splice = pbuffer.splice[:0]
}

//Len 数据长度
func (pbuffer *MyBuffer) Len() int {
	return len(pbuffer.splice)
}

//Append 添加数据
func (pbuffer *MyBuffer) Append(data ...byte) {
	pbuffer.splice = append(pbuffer.splice, data...)
}

//AppendUint8 添加数据
func (pbuffer *MyBuffer) AppendUint8(data uint8) {
	pbuffer.Append(data)
}

//AppendInt8 添加数据
func (pbuffer *MyBuffer) AppendInt8(data int8) {
	pbuffer.AppendUint8(uint8(data))
}

//AppendUint16 添加数据
func (pbuffer *MyBuffer) AppendUint16(data uint16) {
	var arr [2]byte
	binary.BigEndian.PutUint16(arr[:], data)
	pbuffer.splice = append(pbuffer.splice, arr[:]...)
}

//AppendInt16 添加数据
func (pbuffer *MyBuffer) AppendInt16(data int16) {
	pbuffer.AppendUint16(uint16(data))
}

//AppendUint32 添加数据
func (pbuffer *MyBuffer) AppendUint32(data uint32) {
	var arr [4]byte
	binary.BigEndian.PutUint32(arr[:], data)
	pbuffer.splice = append(pbuffer.splice, arr[:]...)
}

//AppendInt32 添加数据
func (pbuffer *MyBuffer) AppendInt32(data int32) {
	pbuffer.AppendUint32(uint32(data))
}

//AppendUint64 添加数据
func (pbuffer *MyBuffer) AppendUint64(data uint64) {
	var arr [8]byte
	binary.BigEndian.PutUint64(arr[:], data)
	pbuffer.splice = append(pbuffer.splice, arr[:]...)
}

//AppendInt64 添加数据
func (pbuffer *MyBuffer) AppendInt64(data int64) {
	pbuffer.AppendUint64(uint64(data))
}

//AppendString 添加数据
func (pbuffer *MyBuffer) AppendString(data string) {
	for i := 0; i < len(data); i++ {
		pbuffer.AppendUint8(data[i])
	}
	pbuffer.AppendUint8(0)
}
