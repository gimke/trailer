package main

import (
	"errors"
	"fmt"
	"github.com/gimke/cartlog"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	name    = "trailer"
	version = "1.0.2"
	reset   = "\033[0m"
	red     = "\033[31;1m"
	green   = "\033[32;1m"
	format  = "%-40s%s\n"
	success = green + "[  OK  ]" + reset
	failed  = red + "[FAILED]" + reset
)

var (
	binPath string
	//flags
	startFlag   bool
	stopFlag    bool
	restartFlag bool
	listFlag    bool
	versionFlag bool
	daemonFlag  bool
	Quit        = make(chan bool)

	Logger = cartlog.GetLogger()

	ErrAlreadyRunning = errors.New("Service is already running")
	ErrAlreadyStopped = errors.New("Service has already been stopped")
	ErrLoadService    = errors.New("Service not exist")
)

func init() {
	bin := filepath.Base(os.Args[0])
	dir := ""
	if bin == name {
		//exec
		dir = filepath.Dir(os.Args[0])
	} else {
		//go run
		dir, _ = os.Getwd()
	}
	binPath, _ = filepath.Abs(dir)
	cartlog.FileSystem("./logs/" + name)
	initService()
}

func initService() {
	file := binPath + "/services/" + name + ".yaml"
	if !isExist(file) {
		os.MkdirAll(binPath+"/services", 0755)
		data := []byte(configText)
		ioutil.WriteFile(file, data, 0666)
	}
}

func usage() {
	fmt.Fprintf(os.Stdout, usageText)
}

func isExist(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func makeFile(path string) *os.File {
	dir := filepath.Dir(path)
	if !isExist(dir) {
		os.MkdirAll(dir, 0755)
	}
	file, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	return file
}

func resovePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	} else {
		if strings.HasPrefix(path, "."+string(os.PathSeparator)) {
			return binPath + path[1:]
		} else {
			return path
		}
	}
}

func printStatus(action, status string, err error) {
	fmt.Printf(format, action, status)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func eraseLine() {
	fmt.Printf("\x1b[%dK", 2) //clear entire line
	fmt.Printf("\r")          //move cursor to beginning of the line
}

const (
	configText = `name: trailer

command:
  - ./trailer
  - -daemon

run_at_load: false
keep_alive: false
`

	usageText = `Usage of trailer:

  -s,-start         Start service
  -q,-stop          Stop service
  -r,-restart       Restart service
  -v,-version       Display version

Usage of trailer console:

  list              List all service
  start             Start a service
  stop              Stop a service
  restart           Restart a service
`
)
