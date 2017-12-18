package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/gimke/cart/logger"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	name    = "trailer"
	reset   = "\033[0m"
	red     = "\033[31;1m"
	green   = "\033[32;1m"
	format  = "%-40s%s\n"
	success = green + "[  OK  ]" + reset
	failed  = red + "[FAILED]" + reset
)

var (
	VERSION = "1.0.0"
	BINDIR  string
	//flags
	startFlag   bool
	stopFlag    bool
	restartFlag bool
	listFlag    bool
	versionFlag bool
	daemonFlag  bool
	shouldQuit  = make(chan bool)
	Quit        = make(chan bool)

	Logger = logger.Logger

	ErrAlreadyRunning = errors.New("Service is already running")
	ErrAlreadyStopped = errors.New("Service has already been stopped")
	ErrLoadService    = errors.New("Service not exist")
)

func init() {
	bin, _ := os.Executable()
	realPath, err := os.Readlink(bin)
	if err == nil {
		bin = realPath
	}
	if filepath.Base(bin) == name {
		BINDIR = filepath.Dir(bin)
	} else {
		BINDIR, _ = os.Getwd()
	}
	logger.SetFileOutput(BINDIR + "/logs/trailer")
	initService()
}

func initService() {
	file := BINDIR + "/services/" + name + ".yml"
	if !isExist(file) {
		os.MkdirAll(BINDIR+"/services", 0755)
		ioutil.WriteFile(file, []byte(configText), 0666)
		demoFile := BINDIR + "/services/demo.yml"
		ioutil.WriteFile(demoFile, []byte(demoText), 0666)
	}
}

func usage() {
	fmt.Fprintf(os.Stdout, usageText)
}

func isExist(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func makeFile(path string) *os.File {
	dir := filepath.Dir(path)
	if !isExist(dir) {
		os.MkdirAll(dir, 0755)
	}
	file, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	return file
}

func resovePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	} else {
		if strings.HasPrefix(path, "."+string(os.PathSeparator)) {
			return BINDIR + path[1:]
		} else {
			return BINDIR + "/" + path
		}
	}
}

func resoveCommand(path string) string {
	if filepath.IsAbs(path) {
		return path
	} else {
		if strings.HasPrefix(path, "."+string(os.PathSeparator)) {
			return BINDIR + path[1:]
		} else {
			return path
		}
	}
}

func printStatus(action, status string, err error) {
	fmt.Printf(format, action, status)
	if err != nil {
		fmt.Println("Error:", err)
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
		name := strings.Join(strings.Split(f.Name, "/")[1:], "/")
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

const branch = "branch"
const release = "release"
const latest = "latest"

func versionType(version string) string {
	if version == "latest" {
		return latest
	}
	if strings.Contains(version, ":") {
		return release
	}
	return branch
	//return branch
}

const (
	configText = `#don't delete this file
name: trailer

command:
  - ./trailer
  - -daemon

pid_file: ./run/trailer.pid
grace: true
`
	demoText = `#name: service name
#env:
#  - CART_MODE=release

#command:
#  - ./home/cartdemo/cartdemo

#pid_file: ./home/cartdemo/cartdemo.pid
#std_out_file: ./home/cartdemo/logs/out.log
#std_err_file: ./home/cartdemo/logs/err.log
#grace: true
#run_at_load: false
#keep_alive: false

#deploy:
#  provider: github (only support github gitlab)
#  token: Personal access tokens (visit https://github.com/settings/tokens or https://gitlab.com/profile/personal_access_tokens and generate a new token)
#  repository: repository address (https://github.com/gimke/cartdemo)
#  version: branchName (e.g master), latest release (e.g latest）or a release described in a file (e.g master:filepath/version.txt)
#  payload: payload url when update success

name: demo

command:
  - ping
  - -c
  - 3
  - 192.168.1.1

run_at_load: true
keep_alive: false

`

	usageText = `Usage of trailer:

  -l,-list          List services
                    +--------------------------------------------+
                    |  list all services ./trailer -l            |
                    +--------------------------------------------+

  -s,-start         Start service
                    +--------------------------------------------+
                    |  start normal service ./trailer -s demo    |
                    |  start daemon service ./trailer -s         |
                    +--------------------------------------------+

  -q,-stop          Stop service
                    +--------------------------------------------+
                    |  stop normal service ./trailer -q demo     |
                    |  stop daemon service ./trailer -q          |
                    +--------------------------------------------+

  -r,-restart       Restart service
                    +--------------------------------------------+
                    |  restart normal service ./trailer -r demo  |
                    |  restart daemon service ./trailer -r       |
                    +--------------------------------------------+

  -v,-version       Display version
                    +--------------------------------------------+
                    |  show trailer version ./trailer -v         |
                    +--------------------------------------------+

`
)
