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
		s := fromFile(name + ".json")
		if s == nil {
			fmt.Printf("load %s error\n", name)
		}
		pid := s.getPid()
		if pid == 0 {
			err := s.run()
			if err != nil {
				fmt.Printf("start %s error %v\n", name, err)

			} else {
				fmt.Printf("start %s (%d) success\n", name, s.Pid)
			}
		} else {
			fmt.Printf("%s is running\n", name)
		}
		break
	case "stop":
		name := commands[1]
		s := fromFile(name + ".json")
		if s == nil {
			fmt.Printf("load %s error\n", name)
		}
		pid := s.getPid()
		if pid != 0 {
			err := s.stop()
			if err != nil {
				fmt.Printf("stop %s error %v\n", name, err)
			} else {
				fmt.Printf("stop %s (%d) success\n", name, pid)
			}
		} else {
			fmt.Printf("%s is not running\n", name)
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
