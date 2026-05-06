package workspace

import (
	"os"
	"os/exec"
	"strings"
)

type Workspace struct {
	RootPath      string
	Config        Config
	History       History
	Metadata      Metadata
	CurrentBranch string
}

func Load(workspacePath string) Workspace {
	w := Workspace{
		RootPath: workspacePath,
	}
	w.CurrentBranch = w.GetCurrentBranch()
	return w
}

func AbsToTildePath(path string) string {
	home, _ := os.UserHomeDir()

	if strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}

	return path
}

func (w *Workspace) GetCurrentBranch() string {
	if w.CurrentBranch != "" {
		return w.CurrentBranch
	}

	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = w.RootPath
	output, err := cmd.Output()
	if err != nil {
		return "main"
	}

	w.CurrentBranch = strings.TrimSpace(string(output))
	return w.CurrentBranch
}

func (w *Workspace) HasUncommittedChanges() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = w.RootPath
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return len(strings.TrimSpace(string(output))) > 0
}
