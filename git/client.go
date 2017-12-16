package git

type Client interface {
	Request(method, url string) (string, error)
	GetConfigFile(branch string) (string, error)
	GetRelease(release string) (string, string, error)
	GetBranch(branch string) (string, string, error)
	DownloadFile(file, url string) error
	Termination()
}
