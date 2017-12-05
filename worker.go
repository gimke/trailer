package main

import (
	"time"
	"log"
)

func Do() {
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