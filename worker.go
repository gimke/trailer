package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type Services []*Service

type Service struct {
	Name      string
	IsRunning bool
	Pid       int
	Config    *Config
}

type Config struct {
	Name      string
	Command   []string
	KeepAlive bool
}

var wg sync.WaitGroup

func Do() {
	services := NewServices()
	services.GetList()
	for _, service := range *services {
		wg.Add(1)
		go service.monitor()
	}
	wg.Wait()
	Quit <- true
}

func NewServices() *Services {
	return &Services{}
}

func (this *Services) GetList() {
	files, err := ioutil.ReadDir(BinaryDir + "/services")
	if err == nil {
		for _, file := range files {
			name := file.Name()
			s := fromFile(name)
			if s != nil {
				*this = append(*this, s)
			}
		}
	}
}

func fromFile(path string) *Service {
	name := path
	if filepath.Ext(name) == ".json" {
		c, err := ioutil.ReadFile(BinaryDir + "/services/" + name)
		if err == nil {
			var config = &Config{}
			err = json.Unmarshal(c, &config)
			if err == nil {
				s := &Service{Name: config.Name}
				pid := s.getPid()
				s.Pid = pid
				s.IsRunning = pid != 0
				s.Config = config
				return s
			}
		}
	}
	return nil
}

func (this *Service) monitor() {
	for {
		pid := this.getPid()
		if pid == 0 {
			//not running keepalive
			log.Printf("%s is not running\n", this.Name)
			this.checkAndRun()
		} else {
			process, _ := os.FindProcess(pid)
			err := process.Signal(syscall.Signal(0))
			if err != nil {
				log.Printf("%s (%d) is not running %v\n", this.Name, pid, err)
				this.checkAndRun()
			} else {
				log.Printf("%s (%d) is running\n", this.Name, pid)
			}
		}
		time.Sleep(10 * time.Second)
		if ShouldQuit {
			wg.Done()
			break
		}
	}
}

func (this *Service) checkAndRun() {
	if this.Config.KeepAlive {
		err := this.run()
		if err != nil {
			log.Printf("%s running error %v\n", this.Name, err)
		}
	} else {
		this.deletePid()
		this.Pid = 0
		this.IsRunning = false
	}
}

func (this *Service) run() error {
	cmd := exec.Command(this.Config.Command[0], this.Config.Command[1:]...)
	dir, _ := filepath.Abs(filepath.Dir(this.Config.Command[0]))
	cmd.Dir = dir

	err := cmd.Start()
	if err != nil {
		return err
	} else {
		go func() {
			cmd.Wait()
		}()
		this.setPid(cmd.Process.Pid)
		this.Pid = cmd.Process.Pid
		this.IsRunning = true
	}
	return nil
}

func (this *Service) stop() error {
	cmd := exec.Command("kill", strconv.Itoa(this.Pid))
	dir, _ := filepath.Abs(filepath.Dir(this.Config.Command[0]))
	cmd.Dir = dir

	err := cmd.Start()
	if err != nil {
		return err
	} else {
		go func() {
			cmd.Wait()
		}()
		this.deletePid()
		this.Pid = 0
		this.IsRunning = false
	}
	return nil
}

func (this *Service) getPid() int {
	content, err := ioutil.ReadFile(BinaryDir + "/run/" + this.Name + ".pid")
	if err != nil {
		return 0
	} else {
		pid, _ := strconv.Atoi(string(content))
		process, _ := os.FindProcess(pid)
		err := process.Signal(syscall.Signal(0))
		if err != nil {
			pid = 0
		}
		return pid
	}
}

func (this *Service) setPid(pid int) error {
	p := []byte(strconv.Itoa(pid))
	err := ioutil.WriteFile(BinaryDir+"/run/"+this.Name+".pid", p, 0666)
	if err != nil {
		return err
	}
	return nil
}
func (this *Service) deletePid() error {
	return os.Remove(BinaryDir + "/run/" + this.Name + ".pid")
}
