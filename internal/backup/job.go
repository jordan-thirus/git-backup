package backup

import (
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/go-git/go-git/v5"
	gitcfg "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jordan-thirus/git-backup/config"
)

type JobDefinition struct {
	log             *slog.Logger
	Repo            *git.Repository
	repoPath        string
	repoIsTemporary bool
	def             config.Repository
	trimmedPath     string
}

func New(cfg config.Backup, jobDef config.Repository, log *slog.Logger) (*JobDefinition, error) {
	temp := !cfg.Enabled || jobDef.ArchiveOnly
	trimmedPath := trimRepoPath(jobDef.Path)
	var repoPath string
	var err error
	if temp {
		repoPath, err = os.MkdirTemp("/tmp", jobDef.Name)
		if err != nil {
			return nil, err
		}
	} else {
		repoPath = path.Join(cfg.Folder, trimmedPath)
		err = os.MkdirAll(repoPath, 0755)
		if err != nil {
			return nil, err
		}
	}

	return &JobDefinition{
		repoPath:        repoPath,
		repoIsTemporary: temp,
		trimmedPath:     trimmedPath,
		def:             jobDef,
		log:             log,
	}, nil
}

func (j *JobDefinition) Open() (*git.Repository, error) {

	repo, err := git.PlainOpen(j.repoPath)
	if err == git.ErrRepositoryNotExists {
		repo, err = git.PlainClone(j.repoPath, false, &git.CloneOptions{
			URL: j.def.Path,
		})
		if err != nil {
			return repo, err
		}
	} else if err != nil {
		return repo, err
	}
	err = repo.Fetch(&git.FetchOptions{
		Prune:    true,
		RefSpecs: []gitcfg.RefSpec{},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		j.log.Error("Fetch failed", slog.String("error", err.Error()))
	}
	j.Repo = repo

	return repo, err
}

func (j *JobDefinition) Checkout(refName string) (*plumbing.Reference, error) {
	if j.Repo == nil {
		panic("repository is nil")
	}
	ref, err := j.Repo.Reference(plumbing.ReferenceName(refName), true)
	if err != nil {
		return nil, err
	}

	j.log.Debug("Looking for reference", slog.Any("ref", ref))
	w, err := j.Repo.Worktree()
	if err != nil {
		return ref, err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash: ref.Hash(),
	})
	if err != nil {
		return ref, err
	}

	headRef, err := j.Repo.Head()
	j.log.Debug("Head ref", slog.Any("head", headRef))
	if err != nil {
		return ref, err
	}
	if ref.Hash() != headRef.Hash() {
		err = fmt.Errorf("hash mismatch for ref %s. expected: %s, actual: %s",
			refName, ref.Hash().String(), headRef.Hash().String())
		return ref, err
	}

	return ref, nil
}

func (j *JobDefinition) Clean() error {
	if j.repoIsTemporary {
		return os.RemoveAll(j.repoPath)
	} else {
		j.Repo.Prune(git.PruneOptions{})
	}
	return nil
}

func (job *JobDefinition) Backup() BackupResult {

	if !job.repoIsTemporary {
		_, err := job.Checkout(job.def.Branch)
		if err != nil {
			return job.BuildErrorResult(err, "backup", job.def.Branch)
		} else {
			return job.BuildResult(ResultTypeSuccess, "backup", job.def.Branch)
		}

	}

	return job.BuildResult(ResultTypeSkipped, "backup", "")

}

func (j *JobDefinition) Archive() []BackupResult {
	results := make([]BackupResult, len(j.def.ArchiveRefs))
	archiveManager := GetArchiveManager()

	for i, r := range j.def.ArchiveRefs {
		ref, err := j.Checkout(r)
		if err != nil {
			j.log.Info(fmt.Sprintf("Failed to archive %s:%s", j.trimmedPath, r),
				slog.String("error", err.Error()))
			results[i] = j.BuildErrorResult(err, "archive", r)
			continue
		}

		if ref.Name().IsTag() && archiveManager.checkIfArchiveExists(j.trimmedPath, r) {
			j.log.Info(fmt.Sprintf("Skipping archive of existing ref %s:%s", j.trimmedPath, r))
			results[i] = j.BuildResult(ResultTypeSkipped, "archive", r)
			continue
		}

		err = archiveManager.archive(j.trimmedPath, r, j.repoPath)
		if err != nil {
			j.log.Info(fmt.Sprintf("Failed to archive %s:%s", j.trimmedPath, r),
				slog.String("error", err.Error()))
			results[i] = j.BuildErrorResult(err, "archive", r)
		} else {
			results[i] = j.BuildResult(ResultTypeSuccess, "archive", r)
		}
	}

	return results
}

func (j *JobDefinition) BuildResult(result ResultType, jobType string, ref string) BackupResult {
	return BackupResult{
		Name:    j.def.Name,
		Ref:     ref,
		JobType: jobType,
		Success: result,
		Msg:     "",
	}
}

func (j *JobDefinition) BuildErrorResult(err error, jobType string, ref string) BackupResult {
	return BackupResult{
		Name:    j.def.Name,
		Ref:     ref,
		JobType: jobType,
		Success: ResultTypeFailed,
		Msg:     err.Error(),
	}
}
