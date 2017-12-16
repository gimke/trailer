package git

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

var _ Client = &Github{}

//deployment:
//  type: github (only support github gitlab)
//  token: Personal access tokens (visit https://github.com/settings/tokens and generate a new token)
//  repository: repository address (https://github.com/gimke/cartdemo)
//  version: latest,v1.0.3,master or other branch
//  payload: payload url when update success

type Github struct {
	Token       string
	Repository  string
	Version string
}

func (g *Github) getUrl() string {
	u, _ := url.Parse(g.Repository)
	return u.Scheme + "://api." + u.Host + "/repos" + u.Path
}

func GithubClient(token, repo, version string) Client {
	return &Github{Token: token, Repository: repo, Version: version}
}

func (g *Github) Request(method, url string) (string, error) {
	req, _ := http.NewRequest(method, url, nil)
	if g.Token != "" {
		req.Header.Set("Authorization", "token "+g.Token)
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

func (g *Github) GetConfigFile() (string, error) {
	u := g.getUrl()
	u += "/contents/.trailer.yml"
	if versionType(g.Version) == branch {
		u+="?ref="+url.PathEscape(g.Version)
	}
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

func (g *Github) GetVersion() (string, string, error) {
	if versionType(g.Version) == branch {
		return g.GetBranche()
	} else {
		return g.GetRelease()
	}
}

func (g *Github) GetRelease() (string, string, error) {
	//latest or tag name
	u := g.getUrl()
	tag := g.Version
	if tag == "latest" {
		u += "/releases/" + tag
	} else {
		u += "/releases/tags/" + tag
	}
	data, err := g.Request("GET", u)
	if err != nil {
		return "", "", err
	}
	var jsonData map[string]interface{}
	err = json.Unmarshal([]byte(data), &jsonData)
	if err != nil {
		return "", "", err
	}
	version := jsonData["name"].(string)
	zip := jsonData["zipball_url"].(string)
	return version, zip, nil
}

func (g *Github) GetBranche() (string, string, error) {
	u := g.getUrl()
	branche := g.Version
	u += "/branches/" + branche
	zip := g.getUrl() + "/zipball/" + branche

	data, err := g.Request("GET", u)
	if err != nil {
		return "", "", err
	}
	var jsonData map[string]interface{}
	err = json.Unmarshal([]byte(data), &jsonData)
	if err != nil {
		return "", "", err
	}
	version := jsonData["commit"].(map[string]interface{})["sha"].(string)

	return version, zip, nil

}

var (
	cx     context.Context
	cancel context.CancelFunc
)

func (g *Github) DownloadFile(file, url string) error {
	// Create the file
	dir := filepath.Dir(file)
	os.MkdirAll(dir, 0755)

	// Get the data
	cx, cancel = context.WithCancel(context.Background())
	req, _ := http.NewRequest("GET", url, nil)
	req = req.WithContext(cx)
	if g.Token != "" {
		req.Header.Set("Authorization", "token "+g.Token)
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

func (g *Github) Termination() {
	//Termination download
	if cancel != nil {
		cancel()
	}
}
