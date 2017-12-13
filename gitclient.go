package main

type gitclient interface {
	NewClient(string)
	GetVersion()
	DownloadFile()
}


var _ gitclient = &github{}

type github struct {

}

func (g *github) NewClient(repo string) {

}

func (g *github) GetVersion() {
	println("version")
}

func (g *github) DownloadFile() {

}