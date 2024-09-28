package config

type Configuration struct {
	Archive      Archive
	Backup       Backup
	Repositories []Repository
}

type Backup struct {
	Enabled bool
	Folder  string
}

type Archive struct {
	Enabled bool
	Folder  string
	Format  string
}

type Repository struct {
	Name        string
	Path        string
	Branch      string
	ArchiveOnly bool
	ArchiveRefs []string
	Releases    []string
}

var Default = &Configuration{
	Backup: Backup{
		Enabled: true,
		Folder:  "./backup",
	},
	Archive: Archive{
		Enabled: true,
		Folder:  "./archive",
		Format:  "zip",
	},
	Repositories: make([]Repository, 0),
}

func New() Configuration {
	return *Default
}
