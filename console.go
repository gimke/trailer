package main

import (
	"fmt"
	"strconv"
)

func consoleExec(commands []string) {
	usage := "Usage: list | start | stop | restart | status"
	if len(commands) == 0 {
		fmt.Println(usage)
		commands = append(commands, "list")
	}
	switch commands[0] {
	case "status":
		s := Service{Name: BinaryName}
		thisPid := s.getPID()
		if thisPid == 0 {
			fmt.Printf("%s is %s%s%s\n", BinaryName, red, "STOPPED", reset)
		} else {
			fmt.Printf("%s (%s) is %s%s%s\n", BinaryName, strconv.Itoa(thisPid), green, "RUNNING", reset)
		}
		break
	case "restart":
		if commands[1] == "all" {
			services := NewServices()
			services.GetList()
			for _, s := range *services {
				consoleRestartService(s.Name)
			}
		} else {
			consoleRestartService(commands[1])
		}
		break
	case "start":
		if commands[1] == "all" {
			services := NewServices()
			services.GetList()
			for _, s := range *services {
				consoleStartService(s.Name)
			}
		} else {
			consoleStartService(commands[1])
		}
		break
	case "stop":
		if commands[1] == "all" {
			services := NewServices()
			services.GetList()
			for _, s := range *services {
				consoleStopService(s.Name)
			}
		} else {
			consoleStopService(commands[1])
		}

		break
	case "list":
		fmt.Printf("%-4s %-6s %-20s %-10s %-10s %-10s\n", "Num", "Pid", "Name", "Status", "RunAtLoad", "KeepAlive")
		services := NewServices()
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
	default:
		fmt.Println(usage)
		break
	}
}

func consoleStartService(name string) {
	action := "Starting service " + name + ":"
	s := fromName(name)
	if s == nil {
		printStatus(fmt.Sprintf(format, action, failed), ErrFile)
	} else {
		pid := s.getPID()
		if pid == 0 {
			err := s.run()
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
func consoleStopService(name string) {
	action := "Stopping service " + name + ":"
	s := fromName(name)
	if s == nil {
		printStatus(fmt.Sprintf(format, action, failed), ErrFile)
	} else {
		pid := s.getPID()
		if pid != 0 {
			err := s.stop()
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
func consoleRestartService(name string) {
	action := "Restarting service " + name + ":"
	s := fromName(name)
	if s == nil {
		printStatus(fmt.Sprintf(format, action, failed), ErrFile)
	} else {
		pid := s.getPID()
		if pid == 0 {
			err := s.run()
			if err != nil {
				printStatus(fmt.Sprintf(format, action, failed), err)
			} else {
				printStatus(fmt.Sprintf(format, action, success), err)
			}
		} else {
			//first stop
			err := s.stop()
			if err != nil {
				printStatus(fmt.Sprintf(format, action, failed), err)
			} else {
				err = s.run()
				if err != nil {
					printStatus(fmt.Sprintf(format, action, failed), err)
				} else {
					printStatus(fmt.Sprintf(format, action, success), err)
				}
			}
			//then run
		}
	}
}
