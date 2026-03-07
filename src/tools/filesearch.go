package tools

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
)

type FileSearchInput struct {
	Query string `json:"query"`
	Path  string `json:"path"`
}

type FileSearchOutput struct {
	Matches []grepMatch
}

type match struct {
	File    string `json:"file"`
	Content string `json:"content"`
}

func RunFileSearch(input FileSearchInput) (FileSearchOutput, error) {

	if input.Query == "" {
		return FileSearchOutput{}, errors.New("query cannot be empty")
	}

	cmd := exec.Command("bash", "-c", fmt.Sprintf("rg --files | rg %s", input.Query))

	fmt.Println(fmt.Sprintf("rg --files | rg %s --json", input.Query))

	if input.Path != "" {
		cmd.Dir = input.Path
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()

	matches, err := ParseGrepMatch(stdout.Bytes())

	if err != nil {
		return FileSearchOutput{}, err
	}

	return FileSearchOutput{
		Matches: matches,
	}, nil
}
