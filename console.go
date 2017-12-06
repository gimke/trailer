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
		thisPid := s.getPid()
		if thisPid == 0 {
			fmt.Printf("%s is %s%s%s\n", BinaryName, red, "STOPPED", reset)
		} else {
			fmt.Printf("%s (%s) is %s%s%s\n", BinaryName, strconv.Itoa(thisPid), green, "RUNNING", reset)
		}
		break
	case "restart":
		name := commands[1]
		action := "Restarting service:"
		s := fromFile(name + ".json")
		if s == nil {
			printStatus(fmt.Sprintf(format, action, failed), ErrFile)
		} else {
			pid := s.getPid()
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
		break
	case "start":
		name := commands[1]
		action := "Starting service:"
		s := fromFile(name + ".json")
		if s == nil {
			printStatus(fmt.Sprintf(format, action, failed), ErrFile)
		} else {
			pid := s.getPid()
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
		break
	case "stop":
		name := commands[1]
		action := "Stopping service:"
		s := fromFile(name + ".json")
		if s == nil {
			printStatus(fmt.Sprintf(format, action, failed), ErrFile)
		} else {
			pid := s.getPid()
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
			pid := strconv.Itoa(s.Pid)
			if !s.IsRunning {
				running = "STOPPED"
				color = red
				pid = "-"
			}
			fmt.Printf("%-4s %-6s %-20s %s%-10s%s %s%-10s%s %s%-10s%s\n", strconv.Itoa(index+1), pid, s.Name,
				color, running, reset,
				runAtLoadColor,runAtLoad,reset,
				keepAliveColor,keepAlive,reset)
		}
		break
	default:
		fmt.Println(usage)
		break
	}
}
