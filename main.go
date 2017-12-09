package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func usage() {
	fmt.Fprintf(os.Stdout, `Usage of trailer:
  -s,-start         Start service
  -q,-stop          Stop service
  -r,-restart       Restart service
  -v,-version       Display version

Usage of trailer commons:
  list              List all service
  start             Start a service
  stop              Stop a service
  restart           Restart a service
`)
}
func main() {

	var (
		startFlag   bool
		stopFlag    bool
		restartFlag bool
		versionFlag bool
		daemonFlag  bool
	)

	flag.Usage = usage

	flag.BoolVar(&startFlag, "start", false, startUsage)
	flag.BoolVar(&startFlag, "s", false, startUsage)

	flag.BoolVar(&stopFlag, "stop", false, stopUsage)
	flag.BoolVar(&stopFlag, "q", false, stopUsage)

	flag.BoolVar(&restartFlag, "restart", false, restartUsage)
	flag.BoolVar(&restartFlag, "r", false, restartUsage)

	flag.BoolVar(&versionFlag, "version", false, versionUsage)
	flag.BoolVar(&versionFlag, "v", false, versionUsage)

	flag.BoolVar(&daemonFlag, "daemon", false, daemonUsage)

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
	if daemonFlag {
		p.Work()
		return
	}
	c := console{}
	c.Exec(flag.Args())
}
