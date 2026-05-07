package prompts

import (
	"fmt"
	"strings"
)

const MainSystemPrompt string = `You are an interactive agent that helps users with software engineering tasks inside ZipCode. Use the available tools to read, edit, and search code.

- Prefer dedicated tools over the shell (file read/edit/write, file search, code search).
- Read files before modifying them. Do not propose changes to code you have not read.
- Keep changes minimal and focused on what was asked. No speculative refactors, abstractions, or extra files.
- Confirm with the user before destructive or hard-to-reverse actions.
- Be concise. Lead with the answer or action.
- Use subagent_ tools when a task requires broad investigation or specialist reasoning; otherwise solve directly.
- If a tool call is denied, do not retry the same call — adjust your approach.`

type SkillSummary struct {
	Name        string
	Description string
}

func BuildSystemPrompt(skills []SkillSummary) string {
	if len(skills) == 0 {
		return MainSystemPrompt
	}

	var sb strings.Builder
	sb.WriteString(MainSystemPrompt)
	sb.WriteString("\n\n## Available Skills\n")
	sb.WriteString("Skills are reusable prompt templates registered in this workspace. ")
	sb.WriteString("To use one, call the `invoke_skill` tool with `skill_name` set to the skill name (no leading slash). ")
	sb.WriteString("The resolved skill prompt will be injected as the next user turn; act on it directly.\n\n")
	for _, s := range skills {
		desc := s.Description
		if desc == "" {
			desc = "(no description)"
		}
		sb.WriteString(fmt.Sprintf("- /%s — %s\n", s.Name, desc))
	}
	return sb.String()
}
