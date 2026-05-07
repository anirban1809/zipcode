package skills

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type internalSkillManifest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	PromptTemplate string `json:"prompt_template"`
}

func parseMarkdownSkill(content string) (name, description, body string, err error) {
	if !strings.HasPrefix(content, "---") {
		return "", "", "", errors.New("missing frontmatter")
	}

	rest := strings.TrimPrefix(content, "---")
	rest = strings.TrimLeft(rest, "\r\n")

	end := strings.Index(rest, "\n---")
	if end == -1 {
		return "", "", "", errors.New("unterminated frontmatter")
	}

	frontmatter := rest[:end]
	body = strings.TrimLeft(rest[end+len("\n---"):], "\r\n")

	for _, line := range strings.Split(frontmatter, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		idx := strings.Index(line, ":")
		if idx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		value = strings.Trim(value, `"'`)

		switch key {
		case "name":
			name = value
		case "description":
			description = value
		}
	}

	if name == "" {
		return "", "", "", errors.New("skill missing name")
	}

	return name, description, body, nil
}

func loadInternalSkills(dir string) (map[string]*Skill, error) {
	out := map[string]*Skill{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var m internalSkillManifest
		if err := json.Unmarshal(content, &m); err != nil {
			continue
		}
		if m.Name == "" {
			continue
		}

		out[m.Name] = &Skill{
			Name:           m.Name,
			Description:    m.Description,
			PromptTemplate: m.PromptTemplate,
			Source:         SourceInternal,
			Path:           path,
			Enabled:        true,
		}
	}
	return out, nil
}

func loadMarkdownSkills(dir string, source SkillSource) (map[string]*Skill, error) {
	out := map[string]*Skill{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		name, desc, body, err := parseMarkdownSkill(string(content))
		if err != nil {
			continue
		}

		out[name] = &Skill{
			Name:           name,
			Description:    desc,
			PromptTemplate: body,
			Source:         source,
			Path:           path,
			Enabled:        true,
		}
	}
	return out, nil
}

func LoadAll(internalDir, globalDir, projectDir string) (map[string]*Skill, error) {
	merged := map[string]*Skill{}

	internal, err := loadInternalSkills(internalDir)
	if err != nil {
		return nil, err
	}
	for k, v := range internal {
		merged[k] = v
	}

	global, err := loadMarkdownSkills(globalDir, SourceGlobal)
	if err != nil {
		return nil, err
	}
	for k, v := range global {
		merged[k] = v
	}

	project, err := loadMarkdownSkills(projectDir, SourceProject)
	if err != nil {
		return nil, err
	}
	for k, v := range project {
		merged[k] = v
	}

	return merged, nil
}
