package workspace

import (
	"os"
	"strings"
)

type Workspace struct {
	RootPath string
	Config   Config
	History  History
	Metadata Metadata
}

func Load(workspacePath string) Workspace {
	return Workspace{
		RootPath: workspacePath,
	}
}

func AbsToTildePath(path string) string {
	home, _ := os.UserHomeDir()

	if strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}

	return path
}
