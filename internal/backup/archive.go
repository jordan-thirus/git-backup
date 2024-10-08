package backup

import (
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/jordan-thirus/git-backup/config"
	"github.com/mholt/archiver"
)

type ArchiveManager struct {
	log *slog.Logger
	cfg config.Archive
}

var instance *ArchiveManager

func GetArchiveManager() *ArchiveManager {
	if instance == nil {
		panic("ArchiveManager not initialized")
	}
	return instance
}

func Init(cfg config.Archive, log *slog.Logger) {
	instance = &ArchiveManager{cfg: cfg, log: log}
}
func (a ArchiveManager) checkIfArchiveExists(repo string, ref string) bool {
	archiveFile := path.Join(a.cfg.Folder, repo, fmt.Sprintf("%s.%s", ref, a.cfg.Format))
	_, err := os.Stat(archiveFile)
	return err == nil
}

func (a ArchiveManager) archive(repo string, ref string, dir string) error {
	archiveRefFolder := path.Join(a.cfg.Folder, repo)
	err := os.MkdirAll(archiveRefFolder, 0755)
	if err != nil {
		return err
	}

	archiveFile := path.Join(archiveRefFolder, fmt.Sprintf("%s.%s", ref, a.cfg.Format))
	if a.checkIfArchiveExists(repo, ref) {
		err = os.Remove(archiveFile)
		if err != nil {
			return err
		}
	}

	entries, err := getTopLevelContents(dir)
	if err != nil {
		a.log.Error("Failed to get directory contents", slog.Any("dir", dir), slog.String("error", err.Error()))
		return err
	}

	err = archiver.Archive(entries, archiveFile)
	if err != nil {
		a.log.Error("Failed to archive repository", slog.String("repo", repo), slog.String("ref", repo), slog.String("error", err.Error()))
		return err
	}

	a.log.Info("Archived ref", slog.String("repo", repo), slog.Any("ref", ref))

	return nil
}
