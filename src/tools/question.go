package tools

var QuestionTool = Tool{
	Type: "function",
	Function: ToolFunction{
		Name:        "question",
		Description: "Ask the user a multiple-choice question and wait for them to pick an option. Use this whenever a decision belongs to the user — clarifying ambiguous requirements, choosing between implementation approaches, confirming a destructive or shared-state action — instead of guessing. The question and options render above the input prompt and the tool returns the selected option verbatim.",
		Parameters: JSONSchema{
			Type: "object",
			Properties: map[string]Schema{
				"question": {
					Type:        "string",
					Description: "The question to display to the user. Should be a single complete sentence ending with a question mark.",
				},
				"options": {
					Type:        "array",
					Description: "The list of mutually exclusive choices presented to the user. Must contain at least two options.",
					Items: &Schema{
						Type: "string",
					},
				},
			},
			Required: []string{"question", "options"},
		},
	},
}

type QuestionInput struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
}
