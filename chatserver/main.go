package main

import (
	"chatserver/chat"
	"chatserver/global"
	"log"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
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
	installSignal()
	chat.Start()
}
