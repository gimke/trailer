package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type services []*service

type service struct {
	Name   string
	Config *config
}

type config struct {
	Name    string
	Env     []string
	Command []string
	PidFile string `yaml:"pid_file"`

	StdOutFile string `yaml:"std_out_file"`
	StdErrFile string `yaml:"std_err_file"`
	Grace      bool   `yaml:"grace"`
	RunAtLoad  bool   `yaml:"run_at_load"`
	KeepAlive  bool   `yaml:"keep_alive"`

	Deployment *deployment
}

type deployment struct {
	Type       string `yaml:"type"`
	Token      string `yaml:"token"`
	Repository string `yaml:"repository"`
	Version    string `yaml:"version"`
	Payload    string `yaml:"payload"`
}

func newServices() *services {
	return &services{}
}

func load(name string) *service {
	file := binPath + "/services/" + name + ".yaml"
	if isExist(file) {
		c, _ := ioutil.ReadFile(file)
		var config = &config{}
		err := yaml.Unmarshal(c, &config)
		if err == nil {
			return &service{Name: config.Name, Config: config}
		}
	}
	return nil
}

func (ss *services) GetList() {
	files, err := ioutil.ReadDir(binPath + "/services")
	if err == nil {
		for _, file := range files {
			basename := file.Name()
			if strings.HasPrefix(basename, ".") {
				continue
			}
			name := strings.TrimSuffix(basename, filepath.Ext(basename))
			s := load(name)
			if s != nil {
				*ss = append(*ss, s)
			}
		}
	}
}

func (s *service) Start() error {
	if s.GetPid() != 0 {
		return ErrAlreadyRunning
	}
	command := resovePath(s.Config.Command[0])
	dir, _ := filepath.Abs(filepath.Dir(command))

	cmd := exec.Command(command, s.Config.Command[1:]...)
	if len(s.Config.Env) > 0 {
		cmd.Env = append(os.Environ(), s.Config.Env...)
	}
	cmd.Dir = dir

	if s.Config.StdOutFile != "" {
		out := makeFile(s.Config.StdOutFile)
		cmd.Stdout = out
	} else {
		out := makeFile(binPath + "/logs/" + s.Config.Name + "/stdout.log")
		cmd.Stdout = out
	}
	if s.Config.StdErrFile != "" {
		err := makeFile(s.Config.StdErrFile)
		cmd.Stderr = err
	} else {
		err := makeFile(binPath + "/logs/" + s.Config.Name + "/stderr.log")
		cmd.Stderr = err
	}

	err := cmd.Start()
	if err != nil {
		return err
	} else {
		go func() {
			cmd.Wait()
		}()
		if s.Config.PidFile == "" {
			s.SetPid(cmd.Process.Pid)
		}
	}
	return nil
}

func (s *service) Stop() error {
	pid := s.GetPid()
	if pid == 0 {
		return ErrAlreadyStopped
	} else {
		err := syscall.Kill(pid, syscall.SIGINT)
		if err != nil {
			return err
		}
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
		<-quitStop
		if s.Config.PidFile == "" {
			s.RemovePid()
		}
	}
	return nil
}
func (s *service) Restart() error {
	pid := s.GetPid()
	if pid != 0 {
		if s.Config.Grace {
			err := syscall.Kill(pid, syscall.SIGUSR2)
			if err != nil {
				return err
			}
		} else {
			err := s.Stop()
			if err != nil {
				return err
			} else {
				err = s.Start()
				if err != nil {
					return err
				}
			}
		}
	} else {
		err := s.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *service) IsExist() bool {
	command := resovePath(s.Config.Command[0])
	if _, err := exec.LookPath(command); err == nil {
		return true
	}
	return false
}

func (s *service) GetPid() int {
	content, err := ioutil.ReadFile(s.pidFile())
	if err != nil {
		return 0
	} else {
		pid, _ := strconv.Atoi(string(content))
		if s.processExist(pid) {
			return pid
		} else {
			return 0
		}
	}
}

func (s *service) SetPid(pid int) {
	pidString := []byte(strconv.Itoa(pid))
	os.MkdirAll(filepath.Dir(s.pidFile()), 0755)
	ioutil.WriteFile(s.pidFile(), pidString, 0666)
}

func (s *service) RemovePid() error {
	return os.Remove(s.pidFile())
}

func (s *service) pidFile() string {
	if s.Config != nil && s.Config.PidFile != "" {
		return s.Config.PidFile
	} else {
		return binPath + "/run/" + s.Name + ".pid"
	}
}

func (s *service) processExist(pid int) bool {
	killErr := syscall.Kill(pid, syscall.Signal(0))
	return killErr == nil
}
