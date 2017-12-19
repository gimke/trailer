package main

import (
	"encoding/json"
	"github.com/gimke/trailer/git"
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
	Logger.Info("Load config")
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
	timer := time.NewTimer(60 * time.Second)
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
		deploy := s.Config.Deploy
		if deploy != nil && deploy.Provider != "" {
			var client git.Client
			switch strings.ToLower(deploy.Provider) {
			case "github":
				client = git.GithubClient(deploy.Token, deploy.Repository)
				break
			case "gitlab":
				client = git.GitlabClient(deploy.Token, deploy.Repository)
			}
			if client != nil {
				s.processGit(client)
			}
		}
	}
}

func (s *service) processGit(client git.Client) {
	//get content from remote git
	var (
		preVersion string
		version    string
		asset      string
		doPayload  = true
		err        error
	)

	defer func() {
		if doPayload {
			payloadUrl := s.Config.Deploy.Payload
			if payloadUrl != "" {
				//Payload callback
				data := url.Values{}
				hostName, _ := os.Hostname()
				jsons := map[string]interface{}{
					"hostName": hostName,
					"name":     s.Name,
				}
				if err != nil {
					jsons["event"] = "update"
					jsons["status"] = "failed"
					jsons["error"] = err.Error()
				} else {
					jsons["event"] = "update"
					jsons["status"] = "success"
					jsons["preVersion"] = preVersion
					jsons["version"] = version
				}
				jsonb, _ := json.Marshal(jsons)
				data.Add("event", "update")
				data.Add("payload", string(jsonb))
				resp, err := http.PostForm(payloadUrl, data)
				if err != nil {
					Logger.Error("%s payload:%s error: %v", s.Name, payloadUrl, err)
				} else {
					resultData, _ := ioutil.ReadAll(resp.Body)
					if resp.StatusCode == 200 {
						Logger.Info("%s payload:%s success: %s", s.Name, payloadUrl, string(resultData))
					} else {
						Logger.Error("%s payload:%s error: %s", s.Name, payloadUrl, string(resultData))
					}
				}
			}
		}
	}()
	config := s.Config
	t := versionType(config.Deploy.Version)
	switch t {
	case branch:
		version, asset, err = client.GetBranch(config.Deploy.Version)
		break
	case latest:
		version, asset, err = client.GetRelease(config.Deploy.Version)
		break
	case release:
		arr := strings.Split(config.Deploy.Version, ":")
		version, err = client.GetContentFile(arr[0], strings.Join(arr[1:], ":"))
		version = strings.TrimSpace(version)
		version = strings.Trim(version, "\n")
		version = strings.Trim(version, "\r")

		if err != nil {
			Logger.Error("%s get file error: %v", s.Name, err)
		}
		version, asset, err = client.GetRelease(version)
		break
	}
	if err != nil {
		Logger.Error("%s find version error: %v", s.Name, err)
		return
	}
	Logger.Info("%s find version:%s asset:%s", s.Name, version, asset)

	//check local version
	preVersion = s.GetVersion()
	if preVersion == version {
		Logger.Info("%s preVersion=newVersion=%s", s.Name, version)
		doPayload = false
		return
	}

	//download zip file and unzip
	dir, _ := filepath.Abs(filepath.Dir(config.Command[0]))
	file := BINDIR + "/update/" + s.Name + "/" + version + ".zip"

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
	err = client.DownloadFile(file, asset)
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
	s.SetVersion(version)
	err = s.Restart()
	if err != nil {
		Logger.Error("%s restart service error %v", s.Name, err)
	} else {
		Logger.Info("%s update service success preVersion:%s newVersion:%s", s.Name, preVersion, version)
	}
}
