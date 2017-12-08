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

type process struct {}

func (*process) Start() (string, error) {
	action := "Starting service:"
	s := service{Name: BinaryName}
	if pid := s.GetPID(); pid != 0 {
		return fmt.Sprintf(format, action, failed), ErrAlreadyRunning
	} else {
		cmd := exec.Command(BinaryDir + "/" + BinaryName)
		cmd.Dir = BinaryDir
		err := cmd.Start()
		if err != nil {
			return fmt.Sprintf(format, action, failed), err
		}
		return fmt.Sprintf(format, action, success), err
	}
}

func (*process) Stop() (string, error) {
	action := "Stopping service:"
	s := service{Name: BinaryName}
	if pid := s.GetPID(); pid == 0 {
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
				if pid := s.GetPID(); pid == 0 {
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
		return fmt.Sprintf(format, action, success), err
	}
}

func (this *process) Restart() (string, error) {
	action := "Restarting service:"
	s := service{Name: BinaryName}
	if pid := s.GetPID(); pid == 0 {
		return this.Start()
	} else {
		dir, _ := os.Getwd()
		cmd := exec.Command("kill", "-USR2",strconv.Itoa(pid))
		cmd.Dir = dir
		err := cmd.Start()
		if err != nil {
			return fmt.Sprintf(format, action, failed), err
		}
		return fmt.Sprintf(format, action, success), err
	}
}
//real work
func (*process) Work() {
	s := service{Name: BinaryName}
	if pid := s.GetPID(); pid != 0 {
		fmt.Fprintln(os.Stderr, "\033[31m"+ErrAlreadyRunning.Error()+"\033[0m")
		os.Exit(1)
	} else {
		cartlog.FileSystem("./logs/"+BinaryName)
		pid := []byte(strconv.Itoa(os.Getpid()))
		if _, err := os.Stat(BinaryDir + "/run"); os.IsNotExist(err) {
			os.Mkdir(BinaryDir+"/run", os.ModePerm)
		}
		ioutil.WriteFile(PidFile, pid, 0666)
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)

		go func() {
			for {
				sig := <-sigs
				log.Println("get signal", sig)
				if sig == syscall.SIGUSR2 {
					log.Println(sig)
					Reload = true
					ShouldQuit = true
				} else {
					Reload = false
					ShouldQuit = true
				}
			}
		}()
		go func() {
			log.Println("service is running")
			Do()
		}()
		<-Quit
		//delete pid
		log.Println("service terminated")
		s.DeletePID()
	}
}
