package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

type master string

func (m master) Version() {
	fmt.Println(version)
}

func (m master) Start() {
	action := "Starting service:"
	if pid := m.GetPid(); pid != 0 {
		//is running
		printStatus(action, failed, ErrAlreadyRunning)
	} else {
		//not running
		cmd := exec.Command(binPath+"/"+name, "-daemon")
		cmd.Dir = binPath
		err := cmd.Start()
		if err != nil {
			printStatus(action, failed, err)
		} else {
			printStatus(action, success, nil)
		}
	}
}

func (m master) Stop() {
	action := "Stopping service:"
	if pid := m.GetPid(); pid != 0 {
		//is running
		err := syscall.Kill(pid, syscall.SIGINT)
		if err != nil {
			printStatus(action, failed, err)
			return
		}
		arr := []string{"Stopping service.", "Stopping service..", "Stopping service..."}
		quitStop := make(chan bool)
		go func() {
			i := 0
			for {
				if pid := m.GetPid(); pid == 0 {
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
		printStatus(action, success, nil)
	} else {
		//not running
		printStatus(action, failed, ErrAlreadyStopped)
	}
}

func (m master) Restart() {
}

func (m master) Process() {
	if pid := m.GetPid(); pid != 0 {
		fmt.Fprintln(os.Stderr, "\033[31m"+ErrAlreadyRunning.Error()+"\033[0m")
		os.Exit(1)
	} else {
		Logger.Info("Service started")
		m.SetPid()
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
		m.RemovePid()
	}
}

func (m master) GetPid() int {
	content, err := ioutil.ReadFile(m.pidFile())
	if err != nil {
		return 0
	} else {
		pid, _ := strconv.Atoi(string(content))
		if m.processExist(pid) {
			return pid
		} else {
			//if process not exist delete pid file
			m.RemovePid()
			return 0
		}
	}
}

func (m master) SetPid() {
	pid := []byte(strconv.Itoa(os.Getpid()))
	os.MkdirAll(filepath.Dir(m.pidFile()), os.ModePerm)
	ioutil.WriteFile(m.pidFile(), pid, 0666)
}
func (m master) RemovePid() error {
	return os.Remove(m.pidFile())
}

func (m master) pidFile() string {
	return binPath + "/run/" + string(m) + ".pid"
}

func (m master) processExist(pid int) bool {
	killErr := syscall.Kill(pid, syscall.Signal(0))
	return killErr == nil
}