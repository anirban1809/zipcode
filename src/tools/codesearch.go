package tools

import (
	"errors"
	"fmt"
)

var CodeSearchTool = Tool{
	Type: "function",
	Function: ToolFunction{
		Name:        "code_search",
		Description: "Search inside source files for specific code patterns, symbols, or text. This tool is used to locate functions, classes, variables, imports, and configuration keys.",
		Parameters: JSONSchema{
			Type: "object",
			Properties: map[string]Schema{
				"message": {
					Type:        "string",
					Description: "Explanation of why the code search is needed",
				},
				"query": {
					Type:        "string",
					Description: "Code pattern, symbol, or text to search for",
				},
				"path": {
					Type:        "string",
					Description: "Optional directory path where the search should start",
				},
			},
			Required: []string{
				"message",
				"query",
			},
		},
	},
}

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
