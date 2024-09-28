package backup

import (
	"fmt"
	"log/slog"

	"github.com/jordan-thirus/git-backup/config"
)

func Run(cfg config.Configuration, log *slog.Logger) BackupResults {
	results := BackupResults{Results: make([]BackupResult, 0)}
	for _, r := range cfg.Repositories {
		log.Info(fmt.Sprintf("Processing %s", r.Name))
		job, err := New(cfg.Backup, r, log)
		if err != nil {
			results.Results = append(results.Results, job.BuildErrorResult(err, "init", ""))
			continue
		}
		_, err = job.Open()
		if err != nil {
			results.Results = append(results.Results, job.BuildErrorResult(err, "init", ""))
			continue
		}
		defer job.Clean()

		results.Results = append(results.Results, job.Backup())
		results.Results = append(results.Results, job.Archive()...)
	}
	return results
}
