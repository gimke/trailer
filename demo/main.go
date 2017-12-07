package main

import (
	"log"
	"time"
	"github.com/gimke/cartlog"
	"strconv"
	"os"
	"io/ioutil"
)

func main() {
	quit := make(chan bool)
	l:=cartlog.Log{}
	l.New()
	myenv := os.Getenv("MY_ENV")
	var pid = []byte(strconv.Itoa(os.Getpid()));
	ioutil.WriteFile("./demo.pid", pid, 0666)
	go func() {
		for {
			log.Println(myenv,time.Now())
			time.Sleep(5*time.Second)
		}
	}()
	<-quit
}
