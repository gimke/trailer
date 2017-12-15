package git

import "regexp"

const branch = "branch"
const release = "release"

type Client interface {
	Request(method, url string) (string, error)
	GetConfigFile() (string, error)
	GetVersion() (string, string, error)
	GetRelease() (string, string, error)
	GetBranche() (string, string, error)
	DownloadFile(file, url string) error
	Termination()
}

func versionType(version string) string {
	if version == "latest" {
		return release
	}
	var validTag = regexp.MustCompile(`^v(\d+\.)?(\d+\.)?(\*|\d+)$`)
	if validTag.MatchString(version) {
		return release
	}
	return branch
	//return branch
}