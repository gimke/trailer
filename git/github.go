package git

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"os"
	"io"
	"context"
)

var _ Client = &Github{}

//deployment:
//  type: github (only support github gitlab)
//  token: Personal access tokens (visit https://github.com/settings/tokens and generate a new token)
//  repository: repository address (https://github.com/gimke/cartdemo)
//  version: latest,v1.0.3,master or other branch
//  payload: payload url when update success

type Github struct {
	Token      string
	Repository string
}

func (g *Github) getUrl() string {
	//https://github.com/gimke/cartdemo
	u, _ := url.Parse(g.Repository)
	return u.Scheme + "://api." + u.Host + "/repos" + u.Path
}

func GithubClient(token, repo string) Client {
	return &Github{Token: token, Repository: repo}
}

var cancel context.CancelFunc
var ctx context.Context
func (g *Github) Request(method, url string) (string, error) {
	c := &http.Client{}
	ctx, cancel = context.WithCancel(context.Background())
	req, _ := http.NewRequest(method, url, nil)
	req = req.WithContext(ctx)
	if g.Token != "" {
		req.Header.Set("Authorization", "token "+g.Token)
	}

	var resp *http.Response
	var err error
	success := make(chan bool)

	go func() {
		resp, err = c.Do(req)
		success <- true
	}()

	select {
	case <-success:
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
	case <-ctx.Done():
		return "",ctx.Err()
	}
}

func (g *Github) GetConfig() (string, error) {
	u := g.getUrl()
	u += "/contents/trailer.yaml"
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

func (g *Github) GetRelease(tag string) (string, string, error) {
	//latest or tag name
	u := g.getUrl()
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

func (g *Github) GetBranche(branche string) (string, string, error) {
	u := g.getUrl()
	u += "/branches/" + branche

	zip := g.getUrl()+"/zipball/"+branche

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
func (g *Github) DownloadFile(file, url string) error {
	// Create the file
	dir := filepath.Dir(file)
	os.MkdirAll(dir, 0755)

	// Get the data
	tr := &http.Transport{} // TODO: copy defaults from http.DefaultTransport
	c := &http.Client{Transport: tr}
	req, _ := http.NewRequest("GET", url, nil)
	if g.Token != "" {
		req.Header.Set("Authorization", "token "+g.Token)
	}
	resp, err := c.Do(req)
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
	return nil
}

func (g *Github) Termination() {
	//Termination download
	if cancel != nil {
		cancel()
	}
}
