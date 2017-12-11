package main

import (
	"flag"
)

func main() {
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

	m := master(name)

	if versionFlag {
		m.Version()
		return
	}
	if startFlag {
		m.Start()
		return
	}
	if stopFlag {
		m.Stop()
		return
	}
	if restartFlag {
		m.Restart()
		return
	}
	if daemonFlag {
		m.Process()
		return
	}

}
