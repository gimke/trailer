package main

import (
	"errors"
	"fmt"
)

const (
	VERSION = "1.0.0"
	PID     = "pid"

	startUsage   = "Start service"
	stopUsage    = "Stop service"
	restartUsage = "Restart service"
	versionUsage = "Display version"
)

var (
	start   bool
	stop    bool
	restart bool
	version bool

	BinaryDir string
	Binary    string
	PidFile       string

	ShouldQuit = false
	Quit       = make(chan bool)

	// ErrAlreadyRunning appears if try to start already running service
	ErrAlreadyRunning = errors.New("Service is already running")

	// ErrAlreadyStopped appears if try to stop already stopped service
	ErrAlreadyStopped = errors.New("Service has already been stopped")

	success = "\t\t\t\t\t[\033[32m" + fmt.Sprintf("%-6s", fmt.Sprintf("%4s", "OK")) + "\033[0m]"
	failed  = "\t\t\t\t\t[\033[31m" + fmt.Sprintf("%-6s", fmt.Sprintf("%6s", "FAILED")) + "\033[0m]"
)
