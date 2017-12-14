package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"github.com/gimke/trailer/merge"
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
	PidFile string `yaml:"pid_file,omitempty"`

	StdOutFile string `yaml:"std_out_file,omitempty"`
	StdErrFile string `yaml:"std_err_file,omitempty"`
	Grace      bool   `yaml:"grace,omitempty"`
	RunAtLoad  bool   `yaml:"run_at_load,omitempty"`
	KeepAlive  bool   `yaml:"keep_alive,omitempty"`

	Deployment *deployment `yaml:"deployment,omitempty"`
}

type deployment struct {
	Type       string `yaml:"type,omitempty"`
	Token      string `yaml:"token,omitempty"`
	Repository string `yaml:"repository,omitempty"`
	Version    string `yaml:"version,omitempty"`
	Payload    string `yaml:"payload,omitempty"`
}

func newServices() *services {
	return &services{}
}

func load(name string) *service {
	file := binPath + "/services/" + name + ".yaml"
	if !isExist(file) {
		return nil
	}
	content, _ := ioutil.ReadFile(file)
	var c = &config{}
	err := yaml.Unmarshal(content, &c)
	if err != nil {
		return nil
	}
	//load protect yaml
	protectFile := binPath + "/services/" + name + ".private.yaml"
	if isExist(file) {
		content, _ := ioutil.ReadFile(protectFile)
		var pc = &config{}
		err = yaml.Unmarshal(content, &pc)
		merge.Merge(pc,c)
		c = pc
	}
	return &service{Name: c.Name, Config: c}
}

func (ss *services) GetList() {
	files, err := ioutil.ReadDir(binPath + "/services")
	if err == nil {
		for _, file := range files {
			basename := file.Name()
			if strings.HasPrefix(basename, ".") {
				continue
			}
			if strings.HasSuffix(basename, ".private.yaml") {
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
		quitStop := make(chan bool)
		go func() {
			for {
				if pid := s.GetPid(); pid == 0 {
					quitStop <- true
					break
				}
				time.Sleep(1 * time.Second)
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
func (s *service) GetVersion() string {
	versionPath := binPath + "/run/" + s.Name + ".ver"
	content, err := ioutil.ReadFile(versionPath)
	if err != nil {
		return ""
	}
	return string(content)
}
func (s *service) SetVersion(version string) {
	versionPath := binPath + "/run/" + s.Name + ".ver"
	data := []byte(version)
	os.MkdirAll(filepath.Dir(versionPath), 0755)
	ioutil.WriteFile(versionPath, data, 0666)
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
