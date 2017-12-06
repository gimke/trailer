package main

import (
	"fmt"
	"os"
	"strconv"
)

func consoleExec(commands []string) {
	usage := "Usage: add | remove | list | start | stop | restart | status"
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
	case "quit":
		os.Exit(1)
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
		fmt.Printf("%-4s %-20s %-10v %-10s\n", "Num", "Name", "Status", "Pid")
		services := NewServices()
		services.GetList()
		for index, s := range *services {
			color := green
			running := "RUNNING"
			pid := strconv.Itoa(s.Pid)
			if !s.IsRunning {
				running = "STOPPED"
				color = red
				pid = "-"
			}
			fmt.Printf("%-4s %-20s %s%-10s%s %-10s\n", strconv.Itoa(index+1), s.Name, color, running, reset, pid)
		}
		break
	default:
		fmt.Println(usage)
		break
	}
}
