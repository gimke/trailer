package main

import (
	"errors"
	"fmt"
	"github.com/gimke/cartlog"
	"os"
	"path/filepath"
)

const (
	name         = "trailer"
	version      = "1.0.2"
	reset        = "\033[0m"
	red          = "\033[31;1m"
	green        = "\033[32;1m"
	format       = "%-40s%s\n"
	success      = green + "[  OK  ]" + reset
	failed       = red + "[FAILED]" + reset
	startUsage   = "Start service"
	stopUsage    = "Stop service"
	restartUsage = "Restart service"
	versionUsage = "Display version"
	daemonUsage  = "Daemon service Please run -s start daemon"
)

var (
	binPath string
	//flags
	startFlag   bool
	stopFlag    bool
	restartFlag bool
	versionFlag bool
	daemonFlag  bool
	Quit        = make(chan bool)

	Logger = cartlog.GetLogger()

	ErrAlreadyRunning = errors.New("Service is already running")
	ErrAlreadyStopped = errors.New("Service has already been stopped")
	ErrFile           = errors.New("Load config file error")
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
}

func usage() {
	fmt.Fprintf(os.Stdout, `Usage of trailer:
  -s,-start         Start service
  -q,-stop          Stop service
  -r,-restart       Restart service
  -v,-version       Display version

Usage of trailer console:
  list              List all service
  start             Start a service
  stop              Stop a service
  restart           Restart a service
`)
}

func printStatus(action, status string, err error) {
	fmt.Printf(format, action, status)
	if err != nil {
		fmt.Println(err)
	}
}

func eraseLine() {
	fmt.Printf("\x1b[%dK", 2) //clear entire line
	fmt.Printf("\r")          //move cursor to beginning of the line
}
