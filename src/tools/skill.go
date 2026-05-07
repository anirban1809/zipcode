package tools

var InvokeSkillTool = Tool{
	Type: "function",
	Function: ToolFunction{
		Name:        "invoke_skill",
		Description: "Invoke a registered skill (a reusable prompt template) by name. The skill's resolved template is injected into the conversation as the next user turn. Use this when a registered skill matches the task at hand instead of writing the steps yourself.",
		Parameters: JSONSchema{
			Type: "object",
			Properties: map[string]Schema{
				"skill_name": {
					Type:        "string",
					Description: "The name of the skill to invoke (without the leading slash).",
				},
				"args": {
					Type:        "string",
					Description: "Optional free-form argument string passed to the skill template as {{args}}.",
				},
			},
			Required: []string{"skill_name"},
		},
	},
}
