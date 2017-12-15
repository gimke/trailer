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
)

var _ Client = &Gitlab{}

type Gitlab struct {
	Token       string
	Repository  string
	Version 	string
}

func (g *Gitlab) getUrl() string {
	u, _ := url.Parse(g.Repository)
	return u.Scheme + "://" + u.Host + "/api/v4/projects/" + url.PathEscape(strings.TrimPrefix(u.Path, "/"))
}

func GitlabClient(token, repo, version string) Client {
	return &Gitlab{Token: token, Repository: repo, Version: version}
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

func (g *Gitlab) GetConfigFile() (string, error) {
	u := g.getUrl()
	fmt.Println(u)
	u += "/repository/files/trailer.yaml"
	if versionType(g.Version) == branch {
		u+="?ref="+url.PathEscape(g.Version)
	}
	fmt.Println(u)
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
	fmt.Println(content)
	return content, nil
}

func (g *Gitlab) GetVersion() (string, string, error) {
	if versionType(g.Version) == branch {
		return g.GetBranche()
	} else {
		return g.GetRelease()
	}
}

func (g *Gitlab) GetRelease() (string, string, error) {
	return "", "", nil
}

func (g *Gitlab) GetBranche() (string, string, error) {
	return "", "", nil
}

func (g *Gitlab) DownloadFile(file, url string) error {
	return nil
}

func (g *Gitlab) Termination() {

}
