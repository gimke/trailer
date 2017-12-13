package git

type Client interface {
	NewClient(string)
	GetVersion()
	DownloadFile()
}