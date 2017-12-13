package main

type git interface {
	NewClient(string)
	GetVersion()
	DownloadFile()
}


var _ git = &github{}

type github struct {

}

func (g *github) NewClient(repo string) {

}

func (g *github) GetVersion() {

}

func (g *github) DownloadFile() {

}