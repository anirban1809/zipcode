package tools

var CreatePlanTool = Tool{
	Type: "function",
	Function: ToolFunction{
		Name:        "create_plan",
		Description: "Declare a multi-step execution plan for the current task. Provide an ordered list of step outlines (one sentence each). The runtime will auto-generate the concrete prompt for each step from its outline and the previous step's output, then run them sequentially. Use this for tasks that span several distinct phases (investigate → design → implement → verify) where each phase's work depends on the previous phase's findings. Do not call this if a plan is already active.",
		Parameters: JSONSchema{
			Type: "object",
			Properties: map[string]Schema{
				"title": {
					Type:        "string",
					Description: "Short label naming the overall plan.",
				},
				"steps": {
					Type:        "array",
					Description: "Ordered step outlines. Each outline should describe what the step accomplishes — not the exact prompt to run. The prompt is generated at execution time from accumulated context.",
					Items: &Schema{
						Type: "string",
					},
				},
			},
			Required: []string{"title", "steps"},
		},
	},
}

type CreatePlanInput struct {
	Title string   `json:"title"`
	Steps []string `json:"steps"`
}
