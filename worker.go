package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"strings"
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

			var client gitclient
			switch strings.ToLower(s.Config.Deployment.Type) {
			case "github":
				client = &github{}
				break
			}
			s.processGit(client)

			//remoteConfig, err := s.getRemoteConfig()
			//if err == nil {
			//	remoteVersion := remoteConfig.Deployment.Version
			//	if remoteVersion != s.Config.Deployment.Version || !s.IsExist() {
			//		Logger.Info("%s begin update remote:%s current:%s", s.Name, remoteVersion, s.Config.Deployment.Version)
			//		dir, _ := filepath.Abs(filepath.Dir(resovePath(remoteConfig.Command[0])))
			//		if !s.IsExist() {
			//			os.MkdirAll(dir, 0755)
			//		}
			//		//todo download file from git and unzip then start service
			//	}
			//} else {
			//	Logger.Error("%s update config error %v", s.Name, err)
			//}
		}
	}
}

func (s *service) processGit(client gitclient) {
	client.GetVersion()
}

//func (s *service) getRemoteConfig() (*config, error) {
//	client := &http.Client{}
//	req, _ := http.NewRequest("GET", s.Config.Deployment.ConfigPath, nil)
//	for _, header := range s.Config.Deployment.ConfigHeaders {
//		kv := strings.Split(header, ":")
//		req.Header.Set(strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]))
//	}
//	resp, err := client.Do(req)
//	if resp != nil {
//		defer resp.Body.Close()
//	}
//	if err != nil {
//		return nil, err
//	} else {
//		data, _ := ioutil.ReadAll(resp.Body)
//		if resp.StatusCode == 200 {
//			//success
//			remoteConfig := &config{}
//			err = yaml.Unmarshal(data, &remoteConfig)
//			if err == nil {
//				return remoteConfig, nil
//			} else {
//				return nil, err
//			}
//		} else {
//			return nil, errors.New(string(data))
//		}
//	}
//}
