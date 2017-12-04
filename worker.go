package main

import (
	"time"
	"log"
	"github.com/gimke/cartlog"
)

func Do() {
	l:= cartlog.Log{}
	l.New()
	for {
		log.Printf("OOO\n")
		log.Println("Test")
		time.Sleep(10 * time.Second)
		log.Println("Done")
		if ShouldQuit {
			break
		}
	}
	Quit <- true
}