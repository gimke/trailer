package git

import "net/url"

var _ Client = &Gitlab{}

type Gitlab struct {
	Token      string
	Repository string
}

func (g *Gitlab) getUrl() string {
	u, _ := url.Parse(g.Repository)
	return u.Scheme + "://api." + u.Host + "/repos" + u.Path
}

func GitlabClient(token, repo string) Client {
	return &Gitlab{Token: token, Repository: repo}
}

func (g *Gitlab) Request(method, url string) (string, error) {
	return "",nil
}

func  (g *Gitlab) GetConfig() (string, error) {
	return "",nil

}

func  (g *Gitlab) GetRelease(tag string) (string, string, error) {
	return "","",nil
}

func  (g *Gitlab) GetBranche(branche string) (string, string, error) {
	return "","",nil
}

func  (g *Gitlab) DownloadFile(file, url string) error {
	return nil
}

func  (g *Gitlab) Termination() {

}