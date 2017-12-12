package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type worker struct {
	mu sync.Mutex
	wg sync.WaitGroup
	firstInit bool
	*services
}

func (w *worker) Work() {
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
			time.Sleep(6 * time.Second)
			if sig == syscall.SIGUSR2 {
				w.Load()
			} else {
				Quit <- true
			}
		}
	}()
	go w.Do()
	<-Quit
	Logger.Info("Service terminated")
	s.RemovePid()
}


func (w *worker) Do() {
	w.Load()
	for {
		w.mu.Lock()
		for _, s := range *w.services {
			w.wg.Add(1)
			//first run it
			Logger.Info(s.Name)
			if w.firstInit {
				//s.RunAtLoad()
			}
			go w.Monitor()
		}
		w.firstInit = false
		w.wg.Wait()
		w.mu.Unlock()
	}
}

func (w *worker) Load() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.services = newServices()
	w.services.GetList()
}
func (w *worker) Monitor() {
	time.Sleep(6*time.Second)
	w.wg.Done()
}