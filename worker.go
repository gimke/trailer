package main

import (
	"github.com/gimke/trailer/git"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"
)

type worker struct {
	wg    sync.WaitGroup
	runed bool
	done  chan bool
	*services
}

func (w *worker) Work() {
	w.done = make(chan bool)

	s := load(name)
	if s.GetPid() != 0 {
		Logger.Error(ErrAlreadyRunning.Error())
		os.Exit(1)
	}
	Logger.Info("Service started")
	s.SetPid(os.Getpid())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	go func() {
		for {
			sig := <-sigs
			Logger.Info("Service Get signal %v", sig)
			if sig == syscall.SIGUSR2 {
				w.ReLoad()
			} else {
				w.Quit()
			}
		}
	}()
	for {
		go w.Do()
		quit := <-Quit
		if quit {
			break
		}
	}
	Logger.Info("Service terminated")
	s.RemovePid()
}

func (w *worker) Do() {
	defer func() {
		shouldQuit = make(chan bool)
		w.runed = true
		w.done <- true
	}()
	w.LoadServices()
	Logger.Info("Service load services")
	for _, s := range *w.services {
		w.wg.Add(1)
		//first run it
		if !w.runed {
			s.RunAtLoad()
		}
		go w.Monitor(s)
	}
	w.wg.Wait()
}
func (w *worker) LoadServices() {
	w.services = newServices()
	w.services.GetList()
}
func (w *worker) ReLoad() {
	close(shouldQuit)
	<-w.done
	Quit <- false
}

func (w *worker) Quit() {
	close(shouldQuit)
	<-w.done
	Quit <- true
}
func loop(s *service) {
	start := time.Now()

	s.KeepAlive()
	s.Update()

	latency := time.Now().Sub(start)
	Logger.Info("%s process time %v", s.Name, latency)
	timer := time.NewTimer(10 * time.Second)
	select {
	case <-timer.C:
		go func() {
			loop(s)
		}()
	case <-shouldQuit:
		return
	}
}
func (w *worker) Monitor(s *service) {
	loop(s)
	select {
	case <-shouldQuit:
		w.wg.Done()
		return
	}
}

func (s *service) RunAtLoad() {
	if pid := s.GetPid(); pid == 0 && s.Config.RunAtLoad {
		err := s.Start()
		if err != nil {
			Logger.Error("%s running error %v", s.Name, err)
		} else {
			Logger.Info("%s running success", s.Name)
		}
	}
}

func (s *service) KeepAlive() {
	if pid := s.GetPid(); pid == 0 && s.Config.KeepAlive {
		err := s.Start()
		if err != nil {
			Logger.Error("%s running error %v", s.Name, err)
		}
	}
}

//update
func (s *service) Update() {
	defer func() {
		if err := recover(); err != nil {
			Logger.Error("[Recovery] panic recovered:%s\n%s", err, string(debug.Stack()))
		}
	}()
	if pid := s.GetPid(); pid != 0 || !s.IsExist() {
		if s.Config.Deployment != nil && s.Config.Deployment.Type != "" {
			var client git.Client
			switch strings.ToLower(s.Config.Deployment.Type) {
			case "github":
				client = git.GithubClient(s.Config.Deployment.Token, s.Config.Deployment.Repository)
				break
				//case "gitlab":
				//	client = &git.Gitlab{}
			}
			s.processGit(client)
		}
	}
}

func (s *service) processGit(client git.Client) {
	//get content from remote git
	var (
		preVersion string
		version    string
		zip        string
		doPayload  = true
	)
	c, err := client.GetConfig()
	defer func() {
		if doPayload {
			payload := s.Config.Deployment.Payload
			if s.Config.Deployment.Payload != "" {
				//Payload callback
				data := url.Values{}
				hostName, _ := os.Hostname()
				data.Add("hostName", hostName)
				data.Add("name", s.Name)
				if err != nil {
					data.Add("status", "failed")
					data.Add("error", err.Error())
				} else {
					data.Add("status", "success")
					data.Add("preVersion", preVersion)
					data.Add("version", version)
				}
				resp, err := http.PostForm(payload, data)
				if err != nil {
					Logger.Error("%s payload:%s error: %v", s.Name, payload, err)
				}
				resultData, _ := ioutil.ReadAll(resp.Body)
				if resp.StatusCode == 200 {
					Logger.Info("%s payload:%s success: %s", s.Name, payload, string(resultData))
				} else {
					Logger.Error("%s payload:%s error: %s", s.Name, payload, string(resultData))
				}
			}
		}
	}()
	if err != nil {
		Logger.Error("%s get config error: %v", s.Name, err)
		return
	}

	remoteConfig := &config{}
	err = yaml.Unmarshal([]byte(c), &remoteConfig)
	if err != nil {
		Logger.Error("%s get config error: %v", s.Name, err)
		return
	}
	t := versionType(remoteConfig.Deployment.Version)
	if t == release {
		version, zip, err = client.GetRelease(remoteConfig.Deployment.Version)
	} else if t == branch {
		version, zip, err = client.GetBranche(remoteConfig.Deployment.Version)
	}
	if err != nil {
		Logger.Error("%s find version error: %v", s.Name, err)
		return
	}
	Logger.Info("%s find version:%s zip:%s", s.Name, version, zip)

	//check local version
	preVersion = s.GetVersion()
	if preVersion == version {
		Logger.Info("%s preVersion=newVersion=%s", s.Name, version)
		doPayload = false
		return
	}

	//download zip file and unzip
	dir, _ := filepath.Abs(filepath.Dir(remoteConfig.Command[0]))
	file := dir + "/update/" + version + ".zip"

	//Termination download when shouldQuit close
	var quitLoop = make(chan bool)
	go func() {
		for {
			select {
			case <-quitLoop:
				return
			case <-shouldQuit:
				client.Termination()
				Logger.Info("%s termination download", s.Name)
				return
			}
		}
	}()
	err = client.DownloadFile(file, zip)
	close(quitLoop)

	if err != nil {
		Logger.Error("%s update download error %v", s.Name, err)
		return
	}
	err = unzip(file, dir)
	if err != nil {
		Logger.Error("%s update unzip file error %v", s.Name, err)
		return
	}
	err = s.updateService(c, version)
	if err != nil {
		Logger.Error("%s update service error %v", s.Name, err)
		return
	}
	Logger.Info("%s update service success preVersion:%s newVersion:%s", s.Name, preVersion, version)
}

func (s *service) updateService(content, version string) error {
	p := binPath + "/services/" + s.Name + ".yaml"
	c := []byte(content)
	err := ioutil.WriteFile(p, c, 0666)
	if err != nil {
		return err
	}
	s.SetVersion(version)
	//check if command changes
	rude := false
	tobeupdate := load(s.Name)

	if strings.Join(tobeupdate.Config.Env, "") == strings.Join(s.Config.Env, "") &&
		strings.Join(tobeupdate.Config.Command, "") == strings.Join(s.Config.Command, "") {
		s.Config = tobeupdate.Config
		rude = false
	} else {
		s.Config = tobeupdate.Config
		rude = true
	}
	if rude {
		err := s.Stop()
		if err != nil {
			return err
		} else {
			err = s.Start()
			if err != nil {
				return err
			}
		}
	} else {
		s.Restart()
	}
	return nil
}
