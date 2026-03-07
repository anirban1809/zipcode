package tools

import (
	"errors"
	"fmt"
)

type CodeSearchInput struct {
	Query string `json:"query"`
	Path  string `json:"path"`
}

type CodeSearchOutput struct {
	Matches []grepMatch
}

func RunCodeSearch(input CodeSearchInput) (CodeSearchOutput, error) {

	if input.Query == "" {
		return CodeSearchOutput{}, errors.New("query cannot be empty")
	}

	command := fmt.Sprintf("rg %s -i --json", input.Query)
	result, err := RunBashCommand(command, input.Path)

	if err != nil {
		return CodeSearchOutput{}, err
	}

	matches, err := ParseGrepMatch([]byte(result))

	if err != nil {
		return CodeSearchOutput{}, err
	}

	return CodeSearchOutput{
		Matches: matches,
	}, nil
}
