package main

import (
	"log"
	"time"
	"github.com/gimke/cartlog"
)

func main() {
	quit := make(chan bool)
	l:=cartlog.Log{}
	l.New()
	go func() {
		for {
			log.Println(time.Now())
			time.Sleep(5*time.Second)
		}
	}()
	<-quit
}
