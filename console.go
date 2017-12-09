package main

import (
	"fmt"
	"os"
	"strconv"
)

type console struct{}

func (this *console) Exec(commands []string) {
	usage := "Usage: list | start | stop | restart | status"
	if len(commands) == 0 {
		fmt.Println(usage)
		commands = append(commands, "list")
	}
	switch commands[0] {
	case "status":
		s := service{Name: BinaryName}
		thisPid := s.GetPID()
		if thisPid == 0 {
			fmt.Printf("%s is %s%s%s\n", BinaryName, red, "STOPPED", reset)
		} else {
			fmt.Printf("%s (%s) is %s%s%s\n", BinaryName, strconv.Itoa(thisPid), green, "RUNNING", reset)
		}
		break
	case "restart":
		if commands[1] == "all" {
			ss := newServices()
			ss.GetList()
			for _, s := range *ss {
				this.Restart(s.Name)
			}
		} else {
			this.Restart(commands[1])
		}
		break
	case "start":
		if commands[1] == "all" {
			ss := newServices()
			ss.GetList()
			for _, s := range *ss {
				this.Start(s.Name)
			}
		} else {
			this.Start(commands[1])
		}
		break
	case "stop":
		if commands[1] == "all" {
			ss := newServices()
			ss.GetList()
			for _, s := range *ss {
				this.Stop(s.Name)
			}
		} else {
			this.Stop(commands[1])
		}

		break
	case "list":
		fmt.Printf("%-4s %-6s %-20s %-10s %-10s %-10s\n", "Num", "Pid", "Name", "Status", "RunAtLoad", "KeepAlive")
		services := newServices()
		services.GetList()
		for index, s := range *services {
			color := green
			running := "RUNNING"
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
			pid := strconv.Itoa(s.PID)
			if !s.IsRunning {
				running = "STOPPED"
				color = red
				pid = "-"
			}
			fmt.Printf("%-4s %-6s %-20s %s%-10s%s %s%-10s%s %s%-10s%s\n", strconv.Itoa(index+1), pid, s.Name,
				color, running, reset,
				runAtLoadColor, runAtLoad, reset,
				keepAliveColor, keepAlive, reset)
		}
		break
	case "add":
		break
	case "remove":
		//remove from list
		this.Remove(commands[1])
		//restart service
		break
	default:
		fmt.Println(usage)
		break
	}
}

func (*console) Start(name string) {
	action := "Starting service " + name + ":"
	s := fromName(name)
	if s == nil {
		printStatus(fmt.Sprintf(format, action, failed), ErrFile)
	} else {
		pid := s.GetPID()
		if pid == 0 {
			err := s.Start()
			if err != nil {
				printStatus(fmt.Sprintf(format, action, failed), err)
			} else {
				printStatus(fmt.Sprintf(format, action, success), err)
			}
		} else {
			printStatus(fmt.Sprintf(format, action, failed), ErrAlreadyRunning)
		}
	}
}
func (*console) Stop(name string) {
	action := "Stopping service " + name + ":"
	s := fromName(name)
	if s == nil {
		printStatus(fmt.Sprintf(format, action, failed), ErrFile)
	} else {
		pid := s.GetPID()
		if pid != 0 {
			err := s.Stop()
			if err != nil {
				printStatus(fmt.Sprintf(format, action, failed), err)
			} else {
				printStatus(fmt.Sprintf(format, action, success), nil)
			}
		} else {
			printStatus(fmt.Sprintf(format, action, failed), ErrAlreadyStopped)
		}
	}
}
func (*console) Restart(name string) {
	action := "Restarting service " + name + ":"
	s := fromName(name)
	if s == nil {
		printStatus(fmt.Sprintf(format, action, failed), ErrFile)
	} else {
		pid := s.GetPID()
		if pid == 0 {
			err := s.Start()
			if err != nil {
				printStatus(fmt.Sprintf(format, action, failed), err)
			} else {
				printStatus(fmt.Sprintf(format, action, success), err)
			}
		} else {
			err := s.Restart()
			if err != nil {
				printStatus(fmt.Sprintf(format, action, failed), err)
			} else {
				printStatus(fmt.Sprintf(format, action, success), err)
			}
		}
	}
}

func (*console) Remove(name string) {
	action := "Remove service " + name + ":"
	errYaml := os.Remove(BinaryDir + "/services/" + name + ".yaml")
	errJson := os.Remove(BinaryDir + "/services/" + name + ".json")

	if errYaml != nil && errJson != nil {
		var err error
		if errYaml != nil {
			err = errYaml
		} else {
			err = errJson
		}
		printStatus(fmt.Sprintf(format, action, failed), err)
	} else {
		p := process{}

		p.Restart()
	}
}
