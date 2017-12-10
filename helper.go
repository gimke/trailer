package main

import (
	"errors"
	"fmt"
	"net/http"
	"io"
	"os"
	"path/filepath"
	"archive/zip"
	"strings"
	"io/ioutil"
)

const (
	VERSION = "1.0.2"
	PID     = "pid"

	reset = "\033[0m"
	red   = "\033[31m"
	green = "\033[32m"

	startUsage   = "Start service"
	stopUsage    = "Stop service"
	restartUsage = "Restart service"
	versionUsage = "Display version"
	daemonUsage  = "Daemon service Please run -s start daemon"
)

var (
	BinaryName string
	BinaryDir  string
	PidFile    string

	ShouldQuit = false
	Quit       = make(chan bool)
	Reload     = false
	format     = "%-40s%s"

	// ErrAlreadyRunning appears if try to start already running service
	ErrAlreadyRunning = errors.New("Service is already running")

	// ErrAlreadyStopped appears if try to stop already stopped service
	ErrAlreadyStopped = errors.New("Service has already been stopped")
	ErrFile           = errors.New("Load config file error")

	success = "[\033[32m" + fmt.Sprintf("%-6s", fmt.Sprintf("%4s", "OK")) + "\033[0m]"
	failed  = "[\033[31m" + fmt.Sprintf("%-6s", fmt.Sprintf("%6s", "FAILED")) + "\033[0m]"
)

func printStatus(status string, err error) {
	if err != nil {
		fmt.Println(status, "\nError:", err)
	} else {
		fmt.Println(status)
	}
}

func eraseLine() {
	fmt.Printf("\x1b[%dK", 2) //clear entire line
	fmt.Printf("\r")          //move cursor to beginning of the line
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		//remove first folder
		name := strings.Join(strings.Split(f.Name,"/")[1:],"/")
		path := filepath.Join(dest, name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func makeFile(path string) *os.File {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}
	file, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	return file
}

func downloadFile(file string, url string) (err error) {

	// Create the file
	out := makeFile(file)
	if err != nil  {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		// Writer the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil  {
			return err
		}
	} else {
		data, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	return nil
}