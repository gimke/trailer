package main

import (
	"errors"
	"fmt"
	"github.com/gimke/cartlog"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"archive/zip"
	"io"
	"regexp"
)

const (
	name    = "trailer"
	version = "1.0.2"
	reset   = "\033[0m"
	red     = "\033[31;1m"
	green   = "\033[32;1m"
	format  = "%-40s%s\n"
	success = green + "[  OK  ]" + reset
	failed  = red + "[FAILED]" + reset
)

var (
	binPath string
	//flags
	startFlag   bool
	stopFlag    bool
	restartFlag bool
	listFlag    bool
	versionFlag bool
	daemonFlag  bool
	shouldQuit  = make(chan bool)
	Quit        = make(chan bool)

	Logger = cartlog.FileSystem("./logs/" + name)

	ErrAlreadyRunning = errors.New("Service is already running")
	ErrAlreadyStopped = errors.New("Service has already been stopped")
	ErrLoadService    = errors.New("Service not exist")
)

func init() {
	bin := filepath.Base(os.Args[0])
	dir := ""
	if bin == name {
		//exec
		dir = filepath.Dir(os.Args[0])
	} else {
		//go run
		dir, _ = os.Getwd()
	}
	binPath, _ = filepath.Abs(dir)
	initService()
}

func initService() {
	file := binPath + "/services/" + name + ".yaml"
	if !isExist(file) {
		os.MkdirAll(binPath+"/services", 0755)
		data := []byte(configText)
		ioutil.WriteFile(file, data, 0666)
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
			return binPath + path[1:]
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

const (
	configText = `name: trailer

command:
  - ./trailer
  - -daemon

grace: true
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
