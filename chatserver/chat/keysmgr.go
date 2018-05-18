package chat

import (
	"chatserver/tree"
	"log"
	"sync"
)

//KeysMgr keys管理
type KeysMgr struct {
	mu sync.RWMutex
	pt *tree.Trie
}

var keysmgr KeysMgr

//SetKeys 设置keys
func (mgr *KeysMgr) SetKeys(k []string) {
	mgr.mu.Lock()
	mgr.pt = nil
	mgr.pt = &(tree.Trie{})
	for _, v := range k {
		mgr.pt.AddKey(v)
	}
	mgr.mu.Unlock()
}

//Replace 替换
func (mgr *KeysMgr) Replace(data []uint8, placeHolder uint8) ([]uint8, bool) {
	mgr.mu.RLock()
	d, b := mgr.pt.Replace(data, placeHolder)
	mgr.mu.RUnlock()
	return d, b
}

func initKeysMgr() {
	keys, err := getKeys()
	if err != nil {
		log.Fatalln(err)
	}
	keysmgr.SetKeys(keys)
}
