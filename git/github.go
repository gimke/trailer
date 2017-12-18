package git

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
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
	d          *download
}

func (g *Github) getUrl() string {
	u, _ := url.Parse(g.Repository)
	if u.Host == "github.com" {
		return u.Scheme + "://api." + u.Host + "/repos" + u.Path
	} else {
		return u.Scheme + "://" + u.Host + "/api/v3/repos" + u.Path
	}
}

func GithubClient(token, repo string) Client {
	return &Github{Token: token, Repository: repo}
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

func (g *Github) GetContentFile(branch, file string) (string, error) {
	u := g.getUrl()
	u += "/contents/" + file + "?ref=" + url.PathEscape(branch)
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

func (g *Github) GetRelease(release string) (string, string, error) {
	//latest or tag name
	u := g.getUrl()
	//tag := g.Version
	if release == "latest" {
		u += "/releases/" + release
	} else {
		u += "/releases/tags/" + release
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
	if version == "" {
		version = jsonData["tag_name"].(string)
	}
	asset := jsonData["zipball_url"].(string)
	return version, asset, nil
}

func (g *Github) GetBranch(branch string) (string, string, error) {
	u := g.getUrl()
	//branche := g.Version
	u += "/branches/" + branch
	asset := g.getUrl() + "/zipball/" + branch

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

	return version, asset, nil

}

func (g *Github) DownloadFile(file, url string) error {
	header := "Authorization: token " + g.Token
	g.d = &download{}
	return g.d.downloadFile(header, file, url)
}

func (g *Github) Termination() {
	//Termination download
	if g.d != nil && g.d.cancel != nil {
		g.d.cancel()
	}
}
