package backup

import "testing"

const pruned string = "github.com/jordan-thirus/git-backup"

func TestTrimRepoPath(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
		desc     string
	}{
		{"http://github.com/jordan-thirus/git-backup.git", pruned, "http / .git"},
		{"https://github.com/jordan-thirus/git-backup.git", pruned, "https / .git"},
		{"ssh://github.com/jordan-thirus/git-backup.git", pruned, "ssh / .git"},
		{"git://github.com/jordan-thirus/git-backup.git", pruned, "git / .git"},
		{"ftp://github.com/jordan-thirus/git-backup.git", pruned, "ftp / git"},
		{"ftps://github.com/jordan-thirus/git-backup.git", pruned, "ftps / git"},
		{"http://github.com/jordan-thirus/git-backup", pruned, "no suffix"},
		{"github.com/jordan-thirus/git-backup.git", pruned, "no prefix"},
		{"bad://http://github.com/jordan-thirus/git-backup.git.bak",
			"bad://http://github.com/jordan-thirus/git-backup.git.bak",
			"nested suffix/prefix"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := trimRepoPath(tc.path)
			if tc.expected != actual {
				t.Errorf("Expected: %s, Actual: %s", tc.expected, actual)
			}
		})
	}
}
