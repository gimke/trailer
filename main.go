package main

import (
	"flag"
	"os"
	"path/filepath"
)

func main() {

	var (
		startFlag        bool
		stopFlag         bool
		restartFlag      bool
		versionFlag      bool
		consoleFlag      bool
	)

	flag.BoolVar(&startFlag, "start", false, startUsage)
	flag.BoolVar(&startFlag, "s", false, startUsage)

	flag.BoolVar(&stopFlag, "stop", false, stopUsage)
	flag.BoolVar(&stopFlag, "q", false, stopUsage)

	flag.BoolVar(&restartFlag, "restart", false, restartUsage)
	flag.BoolVar(&restartFlag, "r", false, restartUsage)

	flag.BoolVar(&versionFlag, "version", false, versionUsage)
	flag.BoolVar(&versionFlag, "v", false, versionUsage)

	flag.BoolVar(&consoleFlag, "console", false, consoleUsage)
	flag.BoolVar(&consoleFlag, "c", false, consoleUsage)

	flag.Parse()
	//get bin path
	BinaryName = filepath.Base(os.Args[0])
	BinaryDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	PidFile = BinaryDir + "/run/" + BinaryName + "." + PID

	p := process{}
	if versionFlag {
		printStatus(VERSION, nil)
		return
	}
	if startFlag {
		status, err := p.Start()
		printStatus(status, err)
		return
	}
	if stopFlag {
		status, err := p.Stop()
		printStatus(status, err)
		return
	}
	if restartFlag {
		status, err := p.Restart()
		printStatus(status, err)
		return
	}
	if consoleFlag {
		c := console{}
		c.Exec(flag.Args())
		return
	}
	p.Work()
}
