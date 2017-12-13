package git

type Client interface {
	Request(method, url string) (string, error)
	GetConfig() (string, error)
	GetRelease(tag string) (string, string, error)
	DownloadFile(file, url string) error
	Termination()
}
