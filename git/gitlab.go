package git

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"os"
	"io"
	"path/filepath"
	"context"
)

var _ Client = &Gitlab{}

type Gitlab struct {
	Token       string
	Repository  string
}

func (g *Gitlab) getUrl() string {
	u, _ := url.Parse(g.Repository)
	return u.Scheme + "://" + u.Host + "/api/v4/projects/" + url.PathEscape(strings.TrimPrefix(u.Path, "/"))+"/repository"
}

func GitlabClient(token, repo string) Client {
	return &Gitlab{Token: token, Repository: repo}
}

func (g *Gitlab) Request(method, url string) (string, error) {
	req, _ := http.NewRequest(method, url, nil)
	if g.Token != "" {
		req.Header.Set("PRIVATE-TOKEN", g.Token)
	}
	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	} else {
		data, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode == 200 {
			//success
			if err == nil {
				return string(data), nil
			} else {
				return "", err
			}
		} else {
			return "", errors.New(string(data))
		}
	}
}

func (g *Gitlab) GetConfigFile(branch string) (string, error) {
	u := g.getUrl()
	fmt.Println(u)
	u += "/files/.trailer.yml?ref="+url.PathEscape(branch)
	data, err := g.Request("GET", u)
	if err != nil {
		return "", err
	}
	var jsonData map[string]interface{}
	err = json.Unmarshal([]byte(data), &jsonData)
	if err != nil {
		return "", err
	}
	decode, err := base64.StdEncoding.DecodeString(jsonData["content"].(string))
	content := string(decode)
	return content, nil
}


func (g *Gitlab) GetRelease(release string) (string, string, error) {
	return "", "", nil
}

func (g *Gitlab) GetBranch(branch string) (string, string, error) {
	u := g.getUrl()
	//branche := g.Version
	u += "/branches/" + branch

	data, err := g.Request("GET", u)
	if err != nil {
		return "", "", err
	}
	var jsonData map[string]interface{}
	err = json.Unmarshal([]byte(data), &jsonData)
	if err != nil {
		return "", "", err
	}
	version := jsonData["commit"].(map[string]interface{})["id"].(string)
	asset := g.getUrl() + "/archive.zip?sha=" + version

	return version, asset, nil
}

func (g *Gitlab) DownloadFile(file, url string) error {
	// Create the file
	dir := filepath.Dir(file)
	os.MkdirAll(dir, 0755)

	// Get the data
	cx, cancel = context.WithCancel(context.Background())
	req, _ := http.NewRequest("GET", url, nil)
	req = req.WithContext(cx)
	if g.Token != "" {
		req.Header.Set("PRIVATE-TOKEN", g.Token)
	}

	done := make(chan bool)

	var err error
	var resp *http.Response
	go func() {
		resp, err = http.DefaultClient.Do(req)
		done <- true
	}()

	select {
	case <-done:
		if resp != nil {
			defer resp.Body.Close()
		}
		if err != nil {
			return err
		}

		if resp.StatusCode == 200 {
			// Writer the body to file
			out, err := os.Create(file)
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, resp.Body)
			if err != nil {
				os.Remove(file)
				return err
			}
		} else {
			data, _ := ioutil.ReadAll(resp.Body)
			return errors.New(string(data))
		}
	case <-cx.Done():
		//canceled
		return cx.Err()
	}
	return nil
}

func (g *Gitlab) Termination() {
	//Termination download
	if cancel != nil {
		cancel()
	}
}