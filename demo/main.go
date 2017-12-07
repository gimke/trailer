package main

import (
	"github.com/gimke/cartlog"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

func main() {
	quit := make(chan bool)
	cartlog.FileSystem("./logs")
	logger := cartlog.GetLogger()
	myenv := os.Getenv("MY_ENV")
	var pid = []byte(strconv.Itoa(os.Getpid()))
	ioutil.WriteFile("./demo.pid", pid, 0666)
	go func() {
		for {
			logger.Info("env: %s year: %v", myenv, time.Now().Year())
			logger.Warn("env: %s year: %v", myenv, time.Now().Year())
			logger.Error("env: %s year: %v", myenv, time.Now().Year())
			time.Sleep(5 * time.Second)
		}
	}()
	<-quit
}
