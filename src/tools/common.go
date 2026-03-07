package tools

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
)

type grepEvent struct {
	Type string     `json:"type"`
	Data grepOutput `json:"data"`
}

type grepOutput struct {
	Path struct {
		Text string `json:"text"`
	} `json:"path"`

	Lines struct {
		Text string `json:"text"`
	} `json:"lines"`

	LineNumber int `json:"line_number"`
}

func RunBashCommand(command string, dir string) (string, error) {
	cmd := exec.Command("bash", "-c", command)

	if dir != "" {
		cmd.Dir = dir
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()

	if err != nil {
		return "", err
	}

	return stdout.String(), nil
}

type grepMatch struct {
	File    string `json:"file"`
	Content string `json:"content"`
}

func ParseGrepMatch(output []byte) ([]grepMatch, error) {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	var matches []grepMatch

	for scanner.Scan() {
		line := scanner.Bytes()
		var event grepEvent
		err := json.Unmarshal(line, &event)

		if err != nil {
			continue
		}

		if event.Type != "match" {
			continue
		}

		match := grepMatch{
			File:    event.Data.Path.Text,
			Content: strings.TrimSpace(event.Data.Lines.Text),
		}

		matches = append(matches, match)
	}

	return matches, nil
}
