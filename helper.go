package main

import (
	"errors"
	"fmt"
)

const (
	VERSION = "1.0.2"
	PID     = "pid"

	reset = "\033[0m"
	red   = "\033[31m"
	green = "\033[32m"

	startUsage   = "Start service"
	stopUsage    = "Stop service"
	restartUsage = "Restart service"
	versionUsage = "Display version"
	consoleUsage = "Console"
	daemonUsage  = "Daemon service Please run -s start daemon"
)

var (
	BinaryName string
	BinaryDir  string
	PidFile    string

	ShouldQuit = false
	Quit       = make(chan bool)
	Reload     = false
	format     = "%-40s%s"

	// ErrAlreadyRunning appears if try to start already running service
	ErrAlreadyRunning = errors.New("Service is already running")

	// ErrAlreadyStopped appears if try to stop already stopped service
	ErrAlreadyStopped = errors.New("Service has already been stopped")
	ErrFile           = errors.New("Load config file error")

	success = "[\033[32m" + fmt.Sprintf("%-6s", fmt.Sprintf("%4s", "OK")) + "\033[0m]"
	failed  = "[\033[31m" + fmt.Sprintf("%-6s", fmt.Sprintf("%6s", "FAILED")) + "\033[0m]"
)

func printStatus(status string, err error) {
	if err != nil {
		fmt.Println(status, "\nError:", err)
	} else {
		fmt.Println(status)
	}
}

func eraseLine() {
	fmt.Printf("\x1b[%dK", 2) //clear entire line
	fmt.Printf("\r")          //move cursor to beginning of the line
}
