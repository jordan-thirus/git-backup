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
			j.log.Error("Clone failed", slog.String("error", err.Error()))
			return repo, err
		}
	} else if err != nil {
		j.log.Error("Open failed", slog.String("error", err.Error()))
		return repo, err
	}
	j.Repo = repo
	err = j.fetchOrigin("")

	if err == git.NoErrAlreadyUpToDate {
		err = nil
	}

	return repo, err
}

func (j *JobDefinition) Checkout(refName string, refType string) (*plumbing.Reference, error) {
	if j.Repo == nil {
		panic("repository is nil")
	}

	w, err := j.Repo.Worktree()

	if refType == "tag" {
		revHash, err := j.Repo.ResolveRevision(plumbing.Revision(refName))
		if err != nil {
			j.log.Error("Checkout failed", slog.String("error", err.Error()))
			return nil, err
		}

		newBranch := plumbing.NewBranchReferenceName(fmt.Sprintf("%s-branch", refName))
		j.log.Debug(fmt.Sprintf("Checking out tag %s to branch %s", refName, newBranch))
		err = w.Checkout(&git.CheckoutOptions{
			Hash:   *revHash,
			Branch: newBranch,
			Create: true,
		})
		if err != nil {
			j.log.Error("Checkout failed", slog.String("error", err.Error()))
			return nil, err
		}
	} else if refType == "branch" {
		branchRefName := plumbing.NewBranchReferenceName(refName)
		branchCoOpts := git.CheckoutOptions{
			Branch: plumbing.ReferenceName(branchRefName),
			Force:  true,
		}
		j.log.Debug("Checking out remote branch")
		if err = w.Checkout(&branchCoOpts); err != nil {
			mirrorRemoteBranchRefSpec := fmt.Sprintf("refs/heads/%s:refs/heads/%s", refName, refName)
			err = j.fetchOrigin(mirrorRemoteBranchRefSpec)
			if err != nil {
				j.log.Error("Fetch failed", slog.String("error", err.Error()))
				return nil, err
			}
			err = w.Checkout(&branchCoOpts)
		}
		if err != nil {
			j.log.Error("Checkout failed", slog.String("error", err.Error()))
			return nil, err
		}

		err = w.Pull(&git.PullOptions{RemoteName: "origin"})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			j.log.Error("Pull failed", slog.String("error", err.Error()))
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unknown type %s", refType)
	}
	return j.Repo.Head()
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
		_, err := job.Checkout(job.def.Branch, "branch")
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
		j.log.Debug(fmt.Sprintf("attempting to archive %s", r))

		if r.Type == "tag" && archiveManager.checkIfArchiveExists(j.trimmedPath, r.Name) {
			j.log.Info(fmt.Sprintf("Skipping archive of existing ref %s:%s", j.trimmedPath, r.Name))
			results[i] = j.BuildResult(ResultTypeSkipped, "archive", r.Name)
			continue
		}

		_, err := j.Checkout(r.Name, r.Type)
		if err != nil {
			j.log.Info(fmt.Sprintf("Failed to archive %s:%s", j.trimmedPath, r),
				slog.String("error", err.Error()))
			results[i] = j.BuildErrorResult(err, "archive", r.Name)
			continue
		}

		err = archiveManager.archive(j.trimmedPath, r.Name, j.repoPath)
		if err != nil {
			j.log.Info(fmt.Sprintf("Failed to archive %s:%s", j.trimmedPath, r.Name),
				slog.String("error", err.Error()))
			results[i] = j.BuildErrorResult(err, "archive", r.Name)
		} else {
			results[i] = j.BuildResult(ResultTypeSuccess, "archive", r.Name)
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

func (j *JobDefinition) fetchOrigin(refSpecStr string) error {
	remote, err := j.Repo.Remote("origin")
	if err != nil {
		return err
	}

	var refSpecs []gitcfg.RefSpec
	if refSpecStr != "" {
		refSpecs = []gitcfg.RefSpec{gitcfg.RefSpec(refSpecStr)}
	}

	if err = remote.Fetch(&git.FetchOptions{
		RefSpecs: refSpecs,
	}); err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("fetch origin failed: %v", err)
	}

	return err
}
