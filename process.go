package main

import (
	"fmt"
	"github.com/gimke/cartlog"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func processStart() (string, error) {
	action := "Starting service:"
	s := Service{Name: BinaryName}
	if pid := s.getPID(); pid != 0 {
		return fmt.Sprintf(format, action, failed), ErrAlreadyRunning
	} else {
		cmd := exec.Command(Binary)
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
	s := Service{Name: BinaryName}
	if pid := s.getPID(); pid == 0 {
		return fmt.Sprintf(format, action, failed), ErrAlreadyStopped
	} else {
		quitStop := make(chan bool)
		dir, _ := os.Getwd()
		cmd := exec.Command("kill", strconv.Itoa(pid))
		cmd.Dir = dir
		err := cmd.Start()
		if err != nil {
			return fmt.Sprintf(format, action, failed), err
		}
		arr := []string{"Stopping service.", "Stopping service..", "Stopping service..."}
		go func() {
			i := 0
			for {
				if pid := s.getPID(); pid == 0 {
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
	s := Service{Name: BinaryName}
	if pid := s.getPID(); pid != 0 {
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
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)

		go func() {
			sig := <-sigs
			log.Println("get signal",sig)
			ShouldQuit = true
		}()
		go func() {
			log.Println("service is running")
			Do()
		}()

		<-Quit
		//delete pid
		log.Println("service terminated")
		s.deletePID()
	}
}
