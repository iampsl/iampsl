// 作者：潘胜良

package main

import (
	"fmt"
	"iampsl/public/mybuffer"
	"iampsl/test/msg"
	"strconv"
	"time"
)

func main() {
	pbuffer := new(mybuffer.MyBuffer)
	var test msg.TestMsg
	test.Age = 20
	test.Name = "iampsl"
	test.Score = 100
	test.Pts = make(map[string]msg.Point, 10)
	for i := 0; i < 10; i++ {
		test.Pts[strconv.Itoa(i)] = msg.Point{X: uint32(i), Y: uint32(i)}
	}
	begTime := time.Now()
	for i := 0; i < 1000000; i++ {
		test.Serialize(pbuffer)
		inbuffer := pbuffer.Data()
		var h msg.Head
		_, l := msg.UnSerializeHead(&h, inbuffer)
		var newtest msg.TestMsg
		newtest.UnSerialize(inbuffer[l:])
		pbuffer.Clear()
	}
	fmt.Println(time.Since(begTime).Seconds())
}
