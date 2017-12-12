package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Work() {
	s:=load(name)
	if s.GetPid() !=0 {
		Logger.Error(ErrAlreadyRunning.Error())
		os.Exit(1)
	}
	Logger.Info("Service started")
	s.SetPid(os.Getpid())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	go func() {
		for {
			sig := <-sigs
			Logger.Info("Service Get signal %v", sig)
			time.Sleep(6 * time.Second)
			if sig == syscall.SIGUSR2 {
				Quit <- true
			} else {
				Quit <- true
			}
		}
	}()
	<-Quit
	Logger.Info("Service terminated")
	s.RemovePid()
}