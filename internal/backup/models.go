package backup

type BackupResults struct {
	Results []BackupResult
}

type BackupResult struct {
	Name    string
	Ref     string
	JobType string
	Success ResultType
	Msg     string
}

type ResultType string

const (
	ResultTypeSuccess ResultType = "success"
	ResultTypeFailed  ResultType = "failed"
	ResultTypeSkipped ResultType = "skipped"
)
