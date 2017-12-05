package main

import (
	"os"
	"fmt"
	"os/exec"
	"io/ioutil"
	"time"
	"github.com/gimke/cartlog"
	"strconv"
	"os/signal"
	"syscall"
	"log"
)
func pidExist(pidFile string) bool {
	if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
		return true
	} else {
		return false
	}
}
func processStart() (string, error) {
	action := "Starting service:"
	if pidExist(PidFile) {
		return fmt.Sprintf(format, action, failed), ErrAlreadyRunning
	} else {
		cmd := exec.Command(Binary,"-run")
		cmd.Dir = BinaryDir
		err := cmd.Start()
		if err != nil {
			return fmt.Sprintf(format, action, failed), err
		}
		return fmt.Sprintf(format, action, success), nil
	}
}

func processStop() (string, error) {
	action := "Stopping service:"
	content, err := ioutil.ReadFile(PidFile)
	if err != nil {
		return fmt.Sprintf(format, action, failed), ErrAlreadyStopped
	} else {
		quitStop := make(chan bool)
		pid := string(content)
		dir, _ := os.Getwd()
		cmd := exec.Command("kill", pid)
		cmd.Dir = dir
		err = cmd.Start()
		if err != nil {
			return fmt.Sprintf(format, action, failed), err
		}
		arr := []string{"Stopping service.", "Stopping service..", "Stopping service..."}
		go func() {
			i := 0
			for {
				if !pidExist(PidFile) {
					quitStop <- true
					break
				}
				fmt.Printf(arr[i])
				if i++; i == len(arr) {
					i = 0
				}
				time.Sleep(1 * time.Second)
				eraseLine()
			}
		}()
		<-quitStop
		return fmt.Sprintf(format, action, success), nil
	}
}
//real work
func processWork() {
	if pidExist(PidFile) {
		fmt.Fprintln(os.Stderr, "\033[31m"+ErrAlreadyRunning.Error()+"\033[0m")
		os.Exit(1)
	} else {
		l := cartlog.Log{}
		l.New()
		pid := []byte(strconv.Itoa(os.Getpid()))
		if _, err := os.Stat(PidDir); os.IsNotExist(err) {
			os.Mkdir(PidDir, os.ModePerm)
		}
		ioutil.WriteFile(PidFile, pid, 0666)
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-sigs
			log.Println(sig)
			ShouldQuit = true
		}()
		go func() {
			Do()
		}()

		<-Quit
		//delete pid
		os.Remove(PidFile)
	}
}

