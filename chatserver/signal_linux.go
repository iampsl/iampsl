package main

import (
	"chatserver/global"
	"os"
	"os/signal"
	"syscall"
)

var sigs = make(chan os.Signal, 1)

func installSignal() {
	signal.Notify(sigs, syscall.SIGUSR1)
	go procSignal()
}

func procSignal() {
	for {
		<-sigs
		global.AppLog.SwitchInfo()
	}
}
