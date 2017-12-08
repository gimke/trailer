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

type services []*service

type service struct {
	Name      string
	IsRunning bool
	PID       int
	Loaded    bool
	Config    *config
}

type config struct {
	Name    string
	Env     []string
	Command []string
	PidFile string `json:"pidFile" yaml:"pid_file"`

	StdOutFile string `json:"stdOutFile" yaml:"std_out_file"`
	StdErrFile string `json:"stdErrFile" yaml:"std_err_file"`
	Grace      bool   `json:"grace" yaml:"grace"`
	RunAtLoad  bool   `json:"runAtLoad" yaml:"run_at_load"`
	KeepAlive  bool   `json:"keepAlive" yaml:"keep_alive"`
}

var wg sync.WaitGroup
var firstInit = true
func Do() {
	ss := newServices()
	ss.GetList()
	for _, s := range *ss {
		wg.Add(1)
		//first run it
		if firstInit {
			s.RunAtLoad()
		}
		firstInit = false
		go s.Monitor()
	}
	wg.Wait()
	if Reload {
		ShouldQuit = false
		Reload = false
		Do()
	} else {
		Quit <- true
	}
}

func newServices() *services {
	return &services{}
}

func fromName(name string) *service {
	//check json file or yaml file
	s := fromFile(name + ".json")
	if s == nil {
		s = fromFile(name + ".yaml")
	}
	return s
}

func fromFile(fileName string) *service {
	//check json file or yaml file
	ext := filepath.Ext(fileName)
	if ext == ".json" || ext == ".yaml" {
		if _, err := os.Stat(BinaryDir + "/services/" + fileName); !os.IsNotExist(err) {
			//find json file
			c, err := ioutil.ReadFile(BinaryDir + "/services/" + fileName)
			if err == nil {
				var config = &config{}
				switch ext {
				case ".json":
					err = json.Unmarshal(c, &config)
					break
				case ".yaml":
					err = yaml.Unmarshal(c, &config)
					break
				}
				if err == nil {
					s := &service{Name: config.Name, Config: config}
					pid := s.GetPID()
					s.PID = pid
					s.IsRunning = pid != 0
					return s
				}
			}
		}
	}

	return nil
}

func (this *services) GetList() {
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

func (this *service) Monitor() {
	for {
		pid := this.GetPID()
		if pid == 0 {
			this.KeepAlive()
		}
		time.Sleep(10 * time.Second)
		if ShouldQuit {
			wg.Done()
			break
		}
	}
}

func (this *service) RunAtLoad() {
	if this.Config.RunAtLoad {
		pid := this.GetPID()
		if pid == 0 {
			err := this.Start()
			if err != nil {
				log.Printf("%s running error %v\n", this.Name, err)
			}
		}
	}
}

func (this *service) KeepAlive() {
	if this.Config.KeepAlive {
		err := this.Start()
		if err != nil {
			log.Printf("%s running error %v\n", this.Name, err)
		}
	} else {
		this.DeletePID()
		this.PID = 0
		this.IsRunning = false
	}
}

func (this *service) abs(filePath string) string {
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
func makeFile(path string) *os.File {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}
	file, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	return file
}

func (this *service) Start() error {
	command := this.abs(this.Config.Command[0])
	dir := filepath.Dir(command)

	cmd := exec.Command(command, this.Config.Command[1:]...)
	if len(this.Config.Env) > 0 {
		cmd.Env = append(os.Environ(), this.Config.Env...)
	}
	cmd.Dir = dir

	if this.Config.StdOutFile != "" {
		out := makeFile(this.Config.StdOutFile)
		cmd.Stdout = out
	} else {
		out := makeFile(BinaryDir + "/logs/" + this.Config.Name + "/stdout.log")
		cmd.Stdout = out
	}
	if this.Config.StdErrFile != "" {
		err := makeFile(this.Config.StdErrFile)
		cmd.Stderr = err
	} else {
		err := makeFile(BinaryDir + "/logs/" + this.Config.Name + "/stderr.log")
		cmd.Stderr = err
	}

	err := cmd.Start()
	if err != nil {
		return err
	} else {
		go func() {
			cmd.Wait()
		}()
		this.SetPID(cmd.Process.Pid)
		this.PID = cmd.Process.Pid
		this.IsRunning = true
	}
	return nil
}

func (this *service) Stop() error {
	cmd := exec.Command("kill", strconv.Itoa(this.PID))

	this.DeletePID()
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

func (this *service) Restart() error {
	if this.Config.Grace {
		cmd := exec.Command("kill", "-USR2", strconv.Itoa(this.PID))

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
		err := this.Stop()
		if err != nil {
			return err
		} else {
			err = this.Start()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *service) GetPIDPath() string {
	if this.Config != nil && this.Config.PidFile != "" {
		pidFile := this.abs(this.Config.PidFile)
		return pidFile
	} else {
		return BinaryDir + "/run/" + this.Name + ".pid"
	}
}

func (this *service) GetPID() int {
	content, err := ioutil.ReadFile(this.GetPIDPath())
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

func (this *service) SetPID(pid int) error {
	p := []byte(strconv.Itoa(pid))
	err := ioutil.WriteFile(this.GetPIDPath(), p, 0666)
	if err != nil {
		return err
	}
	return nil
}

func (this *service) DeletePID() error {
	return os.Remove(this.GetPIDPath())
}
