package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"
)

type processMode int

const (
	START processMode = 1 + iota
	STOP
	RESTART
)

type console struct{}

func (c *console) List() {
	fmt.Printf("%-4s %-6s %-20s %-10s %-10s %-10s %-10s\n", "Num", "Pid", "Name", "Status", "RunAtLoad", "KeepAlive", "AutoUpdate")
	ss := newServices()
	ss.GetList()
	for index, s := range *ss {
		color := red
		running := "STOPPED"
		runAtLoad := "N"
		runAtLoadColor := red
		if s.Config.RunAtLoad {
			runAtLoad = "Y"
			runAtLoadColor = green
		}
		keepAlive := "N"
		keepAliveColor := red
		if s.Config.KeepAlive {
			keepAlive = "Y"
			keepAliveColor = green
		}
		autoUpdate := "N"
		autoUpdateColor := red
		if s.Config.Deploy != nil {
			autoUpdate = "Y"
			ver := s.GetVersion()
			if ver != "" {
				if len(ver) > 8 {
					ver = ver[0:5]+"..."
				}
				autoUpdate = "Y" + " " + ver + ""
			}
			autoUpdateColor = green
		}
		pidString := "-"
		if !s.IsExist() {
			running = "NONE"
		}
		if pid := s.GetPid(); pid != 0 {
			running = "RUNNING"
			color = green
			pidString = strconv.Itoa(pid)
		}
		fmt.Printf("%-4s %-6s %-20s %s%-10s%s %s%-10s%s %s%-10s%s %s%-10s%s\n", strconv.Itoa(index+1), pidString, s.Name,
			color, running, reset,
			runAtLoadColor, runAtLoad, reset,
			keepAliveColor, keepAlive, reset,
			autoUpdateColor, autoUpdate, reset)
	}
}

func (c *console) Start() {
	c.process(START)
}

func (c *console) Stop() {
	c.process(STOP)
}

func (c *console) Restart() {
	c.process(RESTART)
}

func (c *console) process(mode processMode) {
	do := func(name string) {
		action := "Check " + name + ":"
		s := load(name)
		if s == nil {
			printStatus(action, failed, ErrLoadService)
			return
		}
		var err error
		switch mode {
		case START:
			action = "Starting " + name + ":"
			err = s.Start()
			break
		case STOP:
			action = "Stopping " + name + ":"
			arr := []string{"Stopping " + s.Name + ".", "Stopping " + s.Name + "..", "Stopping " + s.Name + "..."}
			quitStop := make(chan bool)
			go func() {
				i := 0
				for {
					if pid := s.GetPid(); pid == 0 {
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
			err = s.Stop()
			<-quitStop
			break
		case RESTART:
			action = "Restarting " + name + ":"
			err = s.Restart()
			break
		}
		if err != nil {
			printStatus(action, failed, err)
		} else {
			printStatus(action, success, err)
		}
	}
	command := flag.Arg(0)
	if command == "" {
		do(name)
		return
	}
	if command == "all" {
		ss := newServices()
		ss.GetList()
		for _, s := range *ss {
			if s.Name != name {
				do(s.Name)
			}
		}
	} else {
		do(command)
	}

}

func (c *console) Add() {

}

func (c *console) Remove() {

}
