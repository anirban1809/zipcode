package tools

import (
	"errors"
	"os"
)

var FileReadTool = Tool{
	Type: "function",
	Function: ToolFunction{
		Name:        "file_read",
		Description: "Read the contents of a file from the workspace. Used to inspect code before making modifications. Avoid reading extremely large files unless necessary.",
		Parameters: JSONSchema{
			Type: "object",
			Properties: map[string]Schema{
				"message": {
					Type:        "string",
					Description: "Explanation of why the file needs to be read",
				},
				"path": {
					Type:        "string",
					Description: "Path to the file that should be read",
				},
			},
			Required: []string{
				"message",
				"path",
			},
		},
	},
}

type FileReadInput struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type FileReadOutput struct {
	Content string `json:"content"`
}

func RunFileRead(input FileReadInput) (FileReadOutput, error) {

	if input.Path == "" {
		return FileReadOutput{}, errors.New("file read path cannot be empty")
	}

	result, err := os.ReadFile(input.Path)

	if err != nil {
		return FileReadOutput{}, err
	}

	return FileReadOutput{
		Content: string(result),
	}, nil
}
