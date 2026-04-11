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
						bug_identifier: investigates runtime errors and likely root causes. 
						code_explorer: explores the codebase and provides insights.`,
						Enum: []any{
							"bug_identifier",
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
