package main

import (
	"github.com/gimke/trailer/git"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

type worker struct {
	wg    sync.WaitGroup
	runed bool
	ended chan bool
	*services
}

func (w *worker) Work() {
	if w.ended == nil {
		w.ended = make(chan bool)
	}
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
		shouldQuit = false
		w.runed = true
		w.ended <- true
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
	shouldQuit = true
	<-w.ended
	Quit <- false
}

func (w *worker) Quit() {
	shouldQuit = true
	<-w.ended
	Quit <- true
}
func (w *worker) Monitor(s *service) {
	for {
		start := time.Now()

		s.KeepAlive()
		s.Update()

		latency := time.Now().Sub(start)
		Logger.Info("%s process time %v", s.Name, latency)
		//check shouldQuit
		for i := 0; i < 60; i++ {
			if shouldQuit {
				break
			}
			time.Sleep(time.Second)
		}
		if shouldQuit {
			w.wg.Done()
			break
		}
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
	c, err := client.GetConfig()
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
	arr := strings.Split(remoteConfig.Deployment.Version, ":")
	var (
		version string
		zip     string
	)
	if arr[0] == "release" {
		version, zip, err = client.GetRelease(arr[1])
	} else if arr[0] == "branch" {
		version, zip, err = client.GetRelease(arr[1])
	}
	if err != nil {
		Logger.Error("%s find version error: %v", s.Name, err)
		return
	}
	Logger.Info("%s find version:%s zip:%s", s.Name, version, zip)

	//check local version
	localVersion := s.GetVersion()
	if localVersion == version {
		return
	}

	//download zip file and unzip
	dir, _ := filepath.Abs(filepath.Dir(remoteConfig.Command[0]))
	file := dir + "/update/" + version + ".zip"
	err = client.DownloadFile(file, zip)
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
	Logger.Info("%s update service success new version:%s", s.Name, version)

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
	Logger.Info("%s update success to version:%v", s.Name, version)
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
