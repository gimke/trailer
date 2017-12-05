package main

import (
	"flag"
	"os"
	"path/filepath"
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

	flag.BoolVar(&console, "console", false, consoleUsage)
	flag.BoolVar(&console, "c", false, consoleUsage)

	flag.BoolVar(&runAsService, "run", false, runAsServiceUsage)

	flag.Parse()

	//get bin path
	BinaryName = filepath.Base(os.Args[0])
	BinaryDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	Binary = BinaryDir + "/" + BinaryName
	PidDir = BinaryDir + "/run"
	PidFile = BinaryDir + "/run/" + BinaryName + "." + PID

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
		processStop()
		status, err := processStart()
		printStatus(status, err)
		return
	}
	if console {
		consoleExec(flag.Args())
		return
	}
	if runAsService {
		processWork()
		return
	}
	flag.Usage()
}
