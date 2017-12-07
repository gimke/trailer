package main

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
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
	Env       []string
	Command   []string
	PidFile   string `json:"pidFile" yaml:"pid_file"`
	Grace     bool   `json:"grace" yaml:"grace"`
	RunAtLoad bool   `json:"runAtLoad" yaml:"run_at_load"`
	KeepAlive bool   `json:"keepAlive" yaml:"keep_alive"`
}

var wg sync.WaitGroup

func Do() {
	services := newServices()
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

func newServices() *Services {
	return &Services{}
}

func fromName(name string) *Service {
	//check json file or yaml file
	s := fromFile(name + ".json")
	if s == nil {
		s = fromFile(name + ".yaml")
	}
	return s
}

func fromFile(fileName string) *Service {
	//check json file or yaml file
	ext := filepath.Ext(fileName)
	if ext == ".json" || ext == ".yaml" {
		if _, err := os.Stat(BinaryDir + "/services/" + fileName); !os.IsNotExist(err) {
			//find json file
			c, err := ioutil.ReadFile(BinaryDir + "/services/" + fileName)
			if err == nil {
				var config = &Config{}
				switch ext {
				case ".json":
					err = json.Unmarshal(c, &config)
					break
				case ".yaml":
					err = yaml.Unmarshal(c, &config)
					break
				}
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

func (this *Services) GetList() {
	files, err := ioutil.ReadDir(BinaryDir + "/services")
	if err == nil {
		for _, file := range files {
			s := fromFile(file.Name())
			if s != nil {
				*this = append(*this, s)
			}
		}
	}
}

func (this *Service) monitor() {
	for {
		pid := this.getPID()
		if pid == 0 {
			this.keepAlive()
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

func (this *Service) abs(filePath string) string {
	var command string
	if path.IsAbs(filePath) {
		//if abs
		command = filePath
	} else {
		if strings.Index(filePath, string(os.PathSeparator)) > -1 {
			command = path.Join(BinaryDir, filePath)
		} else {
			command = filePath
		}
	}
	return command
}

func (this *Service) run() error {
	command := this.abs(this.Config.Command[0])
	dir := filepath.Dir(command)

	cmd := exec.Command(command, this.Config.Command[1:]...)
	if len(this.Config.Env) > 0 {
		cmd.Env = append(os.Environ(), this.Config.Env...)
	}
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

	this.deletePID()
	this.PID = 0
	this.IsRunning = false

	err := cmd.Start()
	if err != nil {
		return err
	} else {
		go func() {
			cmd.Wait()
		}()

	}
	return nil
}

func (this *Service) restart() error {
	if this.Config.Grace {
		cmd := exec.Command("kill", "-USR2",strconv.Itoa(this.PID))

		err := cmd.Start()
		if err != nil {
			return err
		} else {
			go func() {
				cmd.Wait()
			}()
		}
		return nil
	} else {
		err := this.stop()
		if err != nil {
			return err
		} else {
			err = this.run()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *Service) getPIDPath() string {
	if this.Config != nil && this.Config.PidFile != "" {
		pidFile := this.abs(this.Config.PidFile)
		return pidFile
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
