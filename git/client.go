package git

type Client interface {
	GetRelease(tag string) (string, error)
	DownloadFile(file string) error
	Termination()
}
