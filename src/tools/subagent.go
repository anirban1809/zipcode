package tools

func LoadSubAgents() Tool {
	return Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        "subagent",
			Description: "Run a registered specialist sub-agent for a bounded task.",
			Parameters: JSONSchema{
				Type: "object",
				Properties: map[string]Schema{
					"agent": {
						Type: "string",
						Description: `Which specialist to run. 
						bug_investigator: investigates runtime errors and likely root causes. 
						refactor_advisor: proposes safe refactoring plans. 
						test_triage: diagnoses failing tests. 
						codebase_summarizer: summarizes architecture and code structure.`,
						Enum: []any{
							"bug_investigator",
							"code_explorer",
						},
					},
					"task": {
						Type:        "string",
						Description: "The task to perform",
					},
					"context": {
						Type:        "string",
						Description: "Scoped context for the sub-agent to be invoked",
					},
				},
				Required: []string{
					"agent", "task",
				},
			},
		},
	}
}

var SubAgentTool = LoadSubAgents()
