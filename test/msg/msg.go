package msg

import (
	"encoding/binary"
	"fmt"
	"iampsl/public/mybuffer"
	"math"
	"reflect"
)

func myserialize(cmdid uint16, pbuffer *mybuffer.MyBuffer, pmsg interface{}) {
	var h = Head{Cmdid: cmdid}
	begLen := pbuffer.Len()
	pbuffer.AppendUint16(h.Size)
	pbuffer.AppendUint16(h.Cmdid)
	serializeValue(pbuffer, reflect.ValueOf(pmsg).Elem())
	endLen := pbuffer.Len()
	splice := pbuffer.Data()
	binary.BigEndian.PutUint16(splice[begLen:begLen+2], uint16(endLen-begLen))
}

func myunserialize(data []byte, pmsg interface{}) (bool, int) {
	return unserializeValue(data, reflect.ValueOf(pmsg).Elem())
}

func unserializeValue(data []byte, v reflect.Value) (bool, int) {
	switch v.Kind() {
	case reflect.Bool:
		if len(data) < 1 {
			return false, 0
		}
		var b = uint8(data[0])
		if b == 1 {
			v.SetBool(true)
			return true, 1
		}
		if b == 0 {
			v.SetBool(false)
			return true, 1
		}
		return false, 0
	case reflect.Int8:
		if len(data) < 1 {
			return false, 0
		}
		var b = int64(data[0])
		v.SetInt(b)
		return true, 1
	case reflect.Int16:
		if len(data) < 2 {
			return false, 0
		}
		b16 := binary.BigEndian.Uint16(data)
		v.SetInt(int64(b16))
		return true, 2
	case reflect.Int32:
		if len(data) < 4 {
			return false, 0
		}
		b32 := binary.BigEndian.Uint32(data)
		v.SetInt(int64(b32))
		return true, 4
	case reflect.Int64:
		if len(data) < 8 {
			return false, 0
		}
		b64 := binary.BigEndian.Uint64(data)
		v.SetInt(int64(b64))
		return true, 8
	case reflect.Uint8:
		if len(data) < 1 {
			return false, 0
		}
		b8 := data[0]
		v.SetUint(uint64(b8))
		return true, 1
	case reflect.Uint16:
		if len(data) < 2 {
			return false, 0
		}
		b16 := binary.BigEndian.Uint16(data)
		v.SetUint(uint64(b16))
		return true, 2
	case reflect.Uint32:
		if len(data) < 4 {
			return false, 0
		}
		b32 := binary.BigEndian.Uint32(data)
		v.SetUint(uint64(b32))
		return true, 4
	case reflect.Uint64:
		if len(data) < 8 {
			return false, 0
		}
		b64 := binary.BigEndian.Uint64(data)
		v.SetUint(b64)
		return true, 8
	case reflect.Float32:
		if len(data) < 4 {
			return false, 0
		}
		b32 := binary.BigEndian.Uint32(data)
		f := math.Float32frombits(b32)
		v.SetFloat(float64(f))
		return true, 4
	case reflect.Float64:
		if len(data) < 8 {
			return false, 0
		}
		b64 := binary.BigEndian.Uint64(data)
		f := math.Float64frombits(b64)
		v.SetFloat(f)
		return true, 8
	case reflect.Slice:
		if len(data) < 4 {
			return false, 0
		}
		sums := int(binary.BigEndian.Uint32(data))
		processByte := 4
		v.SetLen(sums)
		for i := 0; i < sums; i++ {
			b, l := unserializeValue(data[processByte:], v.Index(i))
			if !b {
				return false, 0
			}
			processByte += l
		}
		return true, processByte
	case reflect.String:
		i := 0
		for ; i < len(data); i++ {
			if data[i] == 0 {
				break
			}
		}
		if i == len(data) {
			return false, 0
		}
		v.SetString(string(data[0:i]))
		return true, i + 1
	case reflect.Struct:
		processByte := 0
		for i := 0; i < v.NumField(); i++ {
			b, l := unserializeValue(data[processByte:], v.Field(i))
			if !b {
				return false, 0
			}
			processByte += l
		}
		return true, processByte
	default:
		panic(fmt.Sprintf("%v is not support", v.Type()))
	}
}

func serializeValue(pbuffer *mybuffer.MyBuffer, v reflect.Value) {
	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			pbuffer.AppendUint8(1)
		} else {
			pbuffer.AppendUint8(0)
		}
	case reflect.Int8:
		pbuffer.AppendInt8(int8(v.Int()))
	case reflect.Int16:
		pbuffer.AppendInt16(int16(v.Int()))
	case reflect.Int32:
		pbuffer.AppendInt32(int32(v.Int()))
	case reflect.Int64:
		pbuffer.AppendInt64(v.Int())
	case reflect.Uint8:
		pbuffer.AppendUint8(uint8(v.Uint()))
	case reflect.Uint16:
		pbuffer.AppendUint16(uint16(v.Uint()))
	case reflect.Uint32:
		pbuffer.AppendUint32(uint32(v.Uint()))
	case reflect.Uint64:
		pbuffer.AppendUint64(v.Uint())
	case reflect.Float32:
		f := math.Float32bits(float32(v.Float()))
		pbuffer.AppendUint32(f)
	case reflect.Float64:
		f := math.Float64bits(v.Float())
		pbuffer.AppendUint64(f)
	case reflect.Slice:
		l := v.Len()
		pbuffer.AppendUint32(uint32(l))
		for i := 0; i < l; i++ {
			serializeValue(pbuffer, v.Index(i))
		}
	case reflect.String:
		pbuffer.AppendString(v.String())
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			serializeValue(pbuffer, v.Field(i))
		}
	default:
		panic(fmt.Sprintf("%v is not support", v.Type()))
	}
}

//UnSerializeHead 反系列化消息头
func UnSerializeHead(phead *Head, data []byte) (bool, int) {
	if len(data) < 2 {
		return false, 0
	}
	phead.Size = binary.BigEndian.Uint16(data)
	if len(data[2:]) < 2 {
		return false, 0
	}
	phead.Cmdid = binary.BigEndian.Uint16(data[2:])
	return true, 4
}

const (
	cmdTryPlay uint16 = 10
)

//Head 消息头
type Head struct {
	Size  uint16
	Cmdid uint16
}

//TryPlay 试玩消息
type TryPlay struct {
	LoginType uint8
}

//Serialize 系列化
func (pmsg *TryPlay) Serialize(pbuffer *mybuffer.MyBuffer) {
	myserialize(cmdTryPlay, pbuffer, pmsg)
}

//UnSerialize 反系列化
func (pmsg *TryPlay) UnSerialize(data []byte) bool {
	b, _ := myunserialize(data, pmsg)
	return b
}
