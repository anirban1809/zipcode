package tools

import (
	"errors"
	"os"
)

type FileReadInput struct {
	Path string `json:"path"`
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
