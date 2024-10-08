package backup

import (
	"os"
	"path"
	"strings"
)

func getTopLevelContents(dir string) ([]string, error) {
	file, err := os.Open(dir)
	if err != nil {
		return nil, err
	}

	names, err := file.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	for i, n := range names {
		names[i] = path.Join(dir, n)
	}

	return names, nil
}

func trimRepoPath(repo string) string {
	tmp := strings.TrimSuffix(repo, ".git")
	tmp = strings.TrimPrefix(tmp, "https://")
	tmp = strings.TrimPrefix(tmp, "http://")
	tmp = strings.TrimPrefix(tmp, "ssh://")
	tmp = strings.TrimPrefix(tmp, "git://")
	tmp = strings.TrimPrefix(tmp, "ftp://")
	tmp = strings.TrimPrefix(tmp, "ftps://")
	return tmp
}
