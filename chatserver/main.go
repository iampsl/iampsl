package main

import (
	"chatServer/global"
	"chatServer/socketmgr"
	"chatServer/user"
	"log"
	"os"
	"path/filepath"
)

func init() {
	log.SetFlags(log.Llongfile | log.LstdFlags)
}

func initCwd() {
	err := os.Chdir(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	initCwd()
	if err := global.LoadConfig("config.xml"); err != nil {
		log.Fatalln(err)
	}
	go socketmgr.TimeoutCheck()
	user.Listen()
}
