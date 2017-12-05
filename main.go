package main

import (
	"flag"
	"fmt"
	"github.com/gimke/cartlog"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

func main() {

	flag.BoolVar(&start, "start", false, startUsage)
	flag.BoolVar(&start, "s", false, startUsage)

	flag.BoolVar(&stop, "stop", false, stopUsage)
	flag.BoolVar(&stop, "q", false, stopUsage)

	flag.BoolVar(&restart, "restart", false, restartUsage)
	flag.BoolVar(&restart, "r", false, restartUsage)

	flag.BoolVar(&version, "version", false, versionUsage)
	flag.BoolVar(&version, "v", false, versionUsage)

	flag.Parse()

	//get bin path
	BinaryDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	Binary = BinaryDir + "/trailer"
	PidFile = BinaryDir + "/" + PID

	if version {
		printStatus(VERSION, nil)
		return
	}
	if start {
		status, err := processStart()
		printStatus(status, err)
		return
	}
	if stop {
		status, err := processStop()
		printStatus(status, err)
		return
	}
	if restart {
		status, err := processStop()
		printStatus(status, err)
		status, err = processStart()
		printStatus(status, err)
		return
	}

	startWork()
}

func processStart() (string, error) {
	action := "Starting service:"
	if _, err := os.Stat(PidFile); !os.IsNotExist(err) {
		return fmt.Sprintf(format, action, failed), ErrAlreadyRunning
	} else {
		cmd := exec.Command(Binary)
		cmd.Dir = BinaryDir
		err = cmd.Start()
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
				if _, err := os.Stat(PidFile); os.IsNotExist(err) {
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
func startWork() {
	if _, err := os.Stat(PidFile); !os.IsNotExist(err) {
		log.Fatalln("Service is already running")
	} else {
		l := cartlog.Log{}
		l.New()
		var pid = []byte(strconv.Itoa(os.Getpid()))
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
