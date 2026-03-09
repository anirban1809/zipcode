package tools

import (
	"bytes"
	"errors"
	"os/exec"
	"time"
)

var BashTool = Tool{
	Type: "function",
	Function: ToolFunction{
		Name: "bash_tool",
		Description: `Execute shell commands in the workspace environment.  
Used for inspecting the filesystem, building projects, running tests, or executing scripts.`,
		Parameters: JSONSchema{
			Type: "object",
			Properties: map[string]Schema{
				"message": {
					Type:        "string",
					Description: "A message describing the command being executed by the llm",
				},
				"command": {
					Type:        "string",
					Description: "the command to be executed",
				},
				"working_directory": {
					Type:        "string",
					Description: "The current working directory",
				},
				"timeout_seconds": {
					Description: "seconds after which the command will be terminated automatically",
					Type:        "integer",
				},
			},
			Required: []string{"message", "command"},
		},
	},
}

// ============================
// BASH TOOL
// ============================

// Input

type BashInput struct {
	Message          string `json:"message"`
	Command          string `json:"command"`
	WorkingDirectory string `json:"working_directory,omitempty"`
	TimeoutSeconds   int    `json:"timeout_seconds,omitempty"`
}

// Output

type BashOutput struct {
	ExitCode   int    `json:"exit_code"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	DurationMs int    `json:"duration_ms"`
}

// execution function
func RunBash(input BashInput) (BashOutput, error) {

	if input.Command == "" {
		return BashOutput{}, errors.New("command cannot be empty")
	}

	timeout := input.TimeoutSeconds

	if timeout == 0 {
		timeout = 30
	}

	start := time.Now()

	cmd := exec.Command("bash", "-c", input.Command)

	if input.WorkingDirectory != "" {
		cmd.Dir = input.WorkingDirectory
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return BashOutput{}, err
		}
	}

	duration := time.Since(start)

	return BashOutput{
		ExitCode:   exitCode,
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		DurationMs: int(duration.Milliseconds()),
	}, nil

}
