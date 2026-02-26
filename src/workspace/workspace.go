package workspace

type Workspace struct {
	RootPath string
	Config   Config
	History  History
	Metadata Metadata
}

func Load(workspacePath string) Workspace {
	return Workspace{}
}
