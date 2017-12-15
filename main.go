package main

import (
	"flag"
	"fmt"
)

/*
todo 1,增加gitlab的支持
*/

func main() {
	flag.Usage = usage

	flag.BoolVar(&startFlag, "start", false, "")
	flag.BoolVar(&startFlag, "s", false, "")
	flag.BoolVar(&stopFlag, "stop", false, "")
	flag.BoolVar(&stopFlag, "quit", false, "")
	flag.BoolVar(&stopFlag, "q", false, "")
	flag.BoolVar(&restartFlag, "restart", false, "")
	flag.BoolVar(&restartFlag, "r", false, "")
	flag.BoolVar(&listFlag, "l", false, "")
	flag.BoolVar(&listFlag, "list", false, "")
	flag.BoolVar(&versionFlag, "v", false, "")
	flag.BoolVar(&versionFlag, "version", false, "")
	flag.BoolVar(&daemonFlag, "daemon", false, "")
	flag.Parse()

	c := &console{}

	if versionFlag {
		fmt.Println(version)
		return
	}

	if listFlag {
		c.List()
		return
	}

	if startFlag {
		c.Start()
		return
	}

	if stopFlag {
		c.Stop()
		return
	}

	if restartFlag {
		c.Restart()
		return
	}

	if daemonFlag {
		w := worker{}
		w.Work()
		return
	}
	usage()
}
