package secrets

import "strings"

func RedactForDisplay(content string) string {
	matches := NewDetector().Detect(content)
	redactedContent := content

	for _, match := range matches {
		sliceToReplace := content[match.Start:match.End]
		redactedContent = strings.ReplaceAll(redactedContent, sliceToReplace, "--redacted--")
	}
	return redactedContent
}

// implement later for workspace sandbox safety
type PlaceholderMap map[string]string

func RedactForModel(content string) (redacted string, mapping PlaceholderMap) {
	return "", PlaceholderMap{"": ""}
}
