package agent

import (
	"os"
)

type RepositorySnapshot struct {
	RootFiles       []string            `json:"root_files"`
	TopLevelDirs    []string            `json:"top_level_dirs"`
	Dependencies    map[string][]string `json:"dependencies"`
	BuildIndicators []string            `json:"build_indicators"`
}

func GenerateRepoSnapshot(root string) (*RepositorySnapshot, error) {
	snapshot := &RepositorySnapshot{
		RootFiles:       []string{},
		TopLevelDirs:    []string{},
		Dependencies:    map[string][]string{},
		BuildIndicators: []string{},
	}

	entries, err := os.ReadDir(root)

	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			snapshot.TopLevelDirs = append(snapshot.TopLevelDirs, name+"/")
			continue
		}

		snapshot.RootFiles = append(snapshot.RootFiles, name)

		switch name {

		// Node / TypeScript
		case "package.json":
			snapshot.BuildIndicators = append(snapshot.BuildIndicators, "package.json")

		case "tsconfig.json":
			snapshot.BuildIndicators = append(snapshot.BuildIndicators, "typescript")

		case "vite.config.ts", "vite.config.js":
			snapshot.BuildIndicators = append(snapshot.BuildIndicators, "vite")

		case "next.config.js", "next.config.mjs":
			snapshot.BuildIndicators = append(snapshot.BuildIndicators, "nextjs")

		case "angular.json":
			snapshot.BuildIndicators = append(snapshot.BuildIndicators, "angular")

		// Python
		case "pyproject.toml":
			snapshot.BuildIndicators = append(snapshot.BuildIndicators, "pyproject")

		case "requirements.txt":
			snapshot.BuildIndicators = append(snapshot.BuildIndicators, "requirements")

		case "manage.py":
			snapshot.BuildIndicators = append(snapshot.BuildIndicators, "django")

		// Go
		case "go.mod":
			snapshot.BuildIndicators = append(snapshot.BuildIndicators, "go_module")

		// Infrastructure
		case "Dockerfile":
			snapshot.BuildIndicators = append(snapshot.BuildIndicators, "docker")

		case "docker-compose.yml", "docker-compose.yaml":
			snapshot.BuildIndicators = append(snapshot.BuildIndicators, "docker_compose")

		case "terraform.tf", "main.tf":
			snapshot.BuildIndicators = append(snapshot.BuildIndicators, "terraform")

		case "serverless.yml":
			snapshot.BuildIndicators = append(snapshot.BuildIndicators, "serverless")

		case "aws-cdk.json":
			snapshot.BuildIndicators = append(snapshot.BuildIndicators, "aws_cdk")
		}
	}

	return snapshot, nil
}
