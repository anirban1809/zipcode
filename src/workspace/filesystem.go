package workspace

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

const maxTreeEntries = 200

// FileTree returns a sorted, newline-separated list of files in the workspace,
// honoring .gitignore via `git ls-files`. Returns an empty string if the
// workspace is not a git repo or the command fails. Truncates at maxTreeEntries
// and appends a notice when truncated.
func (w *Workspace) FileTree() string {
	cmd := exec.Command(
		"git", "ls-files",
		"--cached", "--others", "--exclude-standard",
	)
	cmd.Dir = w.RootPath
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	var files []string
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		files = append(files, line)
	}
	sort.Strings(files)

	total := len(files)
	truncated := false
	if total > maxTreeEntries {
		files = files[:maxTreeEntries]
		truncated = true
	}

	result := strings.Join(files, "\n")
	if truncated {
		result += fmt.Sprintf(
			"\n... (%d more files omitted)",
			total-maxTreeEntries,
		)
	}
	return result
}
