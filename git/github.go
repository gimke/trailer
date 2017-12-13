package git

var _ Client = &Github{}

type Github struct {

}

func (g *Github) NewClient(repo string) {

}

func (g *Github) GetVersion() {
	println("version")
}

func (g *Github) DownloadFile() {

}