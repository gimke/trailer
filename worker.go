package main

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Services []*Service

type Service struct {
	Name      string
	IsRunning bool
	PID       int
	Loaded    bool
	Config    *Config
}

type Config struct {
	Name      string
	Command   []string
	PidFile   string `json:"pidFile" yaml:"pid_file"`
	RunAtLoad bool `json:"runAtLoad" yaml:"run_at_load"`
	KeepAlive bool `json:"keepAlive" yaml:"keep_alive"`
}

var wg sync.WaitGroup

func Do() {
	services := NewServices()
	services.GetList()
	for _, service := range *services {
		wg.Add(1)
		//first run it
		service.runAtLoad()
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
			name := strings.Split(file.Name(), ".")[0]
			s := fromFile(name)
			if s != nil {
				*this = append(*this, s)
			}
		}
	}
}

func fromFile(path string) *Service {
	name := path
	//check json file or yaml file
	fileName := BinaryDir + "/services/" + name
	if _, err := os.Stat(fileName + ".json"); !os.IsNotExist(err) {
		//find json file
		c, err := ioutil.ReadFile(fileName + ".json")
		if err == nil {
			var config = &Config{}
			err = json.Unmarshal(c, &config)
			if err == nil {
				s := &Service{Name: config.Name, Config: config}
				pid := s.getPID()
				s.PID = pid
				s.IsRunning = pid != 0
				return s
			}
		}
	} else {
		if _, err := os.Stat(fileName + ".yaml"); !os.IsNotExist(err) {
			//find yaml file
			c, err := ioutil.ReadFile(fileName + ".yaml")
			if err == nil {
				var config = &Config{}
				err = yaml.Unmarshal(c, &config)
				if err == nil {
					s := &Service{Name: config.Name, Config: config}
					pid := s.getPID()
					s.PID = pid
					s.IsRunning = pid != 0
					return s
				}
			}
		}
	}

	return nil
}

func (this *Service) monitor() {
	for {
		pid := this.getPID()
		if pid == 0 {
			//not running keepalive
			log.Printf("%s is not running\n", this.Name)
			this.keepAlive()
		} else {
			log.Printf("%s (%d) is running\n", this.Name, pid)
		}
		time.Sleep(10 * time.Second)
		if ShouldQuit {
			wg.Done()
			break
		}
	}
}
func (this *Service) runAtLoad() {
	if this.Config.RunAtLoad {
		pid := this.getPID()
		if pid == 0 {
			err := this.run()
			if err != nil {
				log.Printf("%s running error %v\n", this.Name, err)
			}
		}
	}
}
func (this *Service) keepAlive() {
	if this.Config.KeepAlive {
		err := this.run()
		if err != nil {
			log.Printf("%s running error %v\n", this.Name, err)
		}
	} else {
		this.deletePID()
		this.PID = 0
		this.IsRunning = false
	}
}

func (this *Service) run() error {
	command, _:=filepath.Abs(this.Config.Command[0])
	dir, _ := filepath.Abs(filepath.Dir(this.Config.Command[0]))
	cmd := exec.Command(command, this.Config.Command[1:]...)
	cmd.Dir = dir

	err := cmd.Start()
	if err != nil {
		return err
	} else {
		go func() {
			cmd.Wait()
		}()
		this.setPID(cmd.Process.Pid)
		this.PID = cmd.Process.Pid
		this.IsRunning = true
	}
	return nil
}

func (this *Service) stop() error {
	cmd := exec.Command("kill", strconv.Itoa(this.PID))
	dir, _ := filepath.Abs(filepath.Dir(this.Config.Command[0]))
	cmd.Dir = dir

	err := cmd.Start()
	if err != nil {
		return err
	} else {
		go func() {
			cmd.Wait()
		}()
		this.deletePID()
		this.PID = 0
		this.IsRunning = false
	}
	return nil
}

func (this *Service) getPIDPath() string {
	if this.Config != nil && this.Config.PidFile != "" {
		return this.Config.PidFile
	} else {
		return BinaryDir + "/run/" + this.Name + ".pid"
	}
}

func (this *Service) getPID() int {
	content, err := ioutil.ReadFile(this.getPIDPath())
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

func (this *Service) setPID(pid int) error {
	p := []byte(strconv.Itoa(pid))
	err := ioutil.WriteFile(this.getPIDPath(), p, 0666)
	if err != nil {
		return err
	}
	return nil
}
func (this *Service) deletePID() error {
	return os.Remove(this.getPIDPath())
}
