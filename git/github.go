package git

import (
	"fmt"
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
}
func (g *Github) getUrl() string {
	//https://github.com/gimke/cartdemo
	u,_ := url.Parse(g.Repository)
	return u.Scheme+"://api."+u.Host+"/repos"+u.Path
}

func GithubClient(token, repo string) Client {
	return &Github{Token:token,Repository:repo}
}

func (g *Github) GetRelease(tag string) (string, error) {
	//latest or tag name
	u := g.getUrl()
	if tag == "latest" {
		u += "/releases/"+tag
	} else {
		u += "/releases/tags/"+tag
	}
	fmt.Println(u)
	return "", nil
}

func (g *Github) DownloadFile(file string) error {
	return nil
}

func (g *Github) Termination() {
	//Termination download
}
