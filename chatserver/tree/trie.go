package tree

import (
	"unicode/utf8"
)

type trieNode struct {
	childs map[uint8]*trieNode
	isend  bool
}

//Trie 前缀树
type Trie struct {
	root trieNode
}

//AddKey 添加键
func (t *Trie) AddKey(k string) {
	length := len(k)
	if length == 0 {
		return
	}
	indexNode := &(t.root)
	for i := 0; i < length; i++ {
		child, ok := indexNode.childs[k[i]]
		if ok {
			indexNode = child
		} else {
			pnew := new(trieNode)
			if indexNode.childs == nil {
				indexNode.childs = make(map[uint8]*trieNode)
			}
			indexNode.childs[k[i]] = pnew
			indexNode = pnew
		}
	}
	indexNode.isend = true
}

func (t *Trie) prefixCmp(data []uint8) int {
	length := len(data)
	if length == 0 {
		return 0
	}
	indexNode := &(t.root)
	for i := 0; i < length; i++ {
		child, ok := indexNode.childs[data[i]]
		if !ok {
			return 0
		}
		if child.isend {
			return i + 1
		}
		indexNode = child
	}
	return 0
}

//Replace 替换
func (t *Trie) Replace(data []uint8, placeHolder uint8) ([]uint8, bool) {
	length := len(data)
	if length == 0 {
		return data, false
	}
	indexSplice := make([]int, 0, 10)
	lengthSplice := make([]int, 0, 10)

	for index := 0; index < length; {
		cmpLen := t.prefixCmp(data[index:])
		if cmpLen != 0 {
			indexSplice = append(indexSplice, index)
			lengthSplice = append(lengthSplice, cmpLen)
			index += cmpLen
		} else {
			_, size := utf8.DecodeRune(data[index:])
			index += size
		}
	}
	if len(indexSplice) == 0 {
		return data, false
	}
	index := indexSplice[0]
	for i := 0; i < len(indexSplice); i++ {
		tmpIndex := indexSplice[i]
		tmpLength := lengthSplice[i]
		nums := utf8.RuneCount(data[tmpIndex : tmpIndex+tmpLength])
		for j := 0; j < nums; j++ {
			data[index] = placeHolder
			index++
		}
		beg := tmpIndex + tmpLength
		if beg < length {
			end := length
			if i+1 < len(indexSplice) {
				end = indexSplice[i+1]
			}
			copy(data[index:], data[beg:end])
			index += (end - beg)
		}
	}
	return data[0:index], true
}
