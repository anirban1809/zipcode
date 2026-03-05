package agent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"zipcode/src/llm/prompts"
	llm "zipcode/src/llm/provider"
	"zipcode/src/workspace"
)

type Plan struct {
	Steps       []PlanStep
	Validations []StepValidationResult
}

type Intent struct {
	Category                 string   `json:"category"`
	OperationType            string   `json:"operation_type"`
	RiskLevel                string   `json:"risk_level"`
	RequiresNewFiles         bool     `json:"requires_new_files"`
	RequiresFileModification bool     `json:"requires_file_modification"`
	RequiresDeletion         bool     `json:"requires_deletion"`
	SearchIdentifiers        []string `json:"search_identifiers"`
	TargetFiles              []any    `json:"target_files"`
}

type PlanStep struct {
	StepId   int
	StepTask string
}

type Planner struct {
	llm       llm.LLMProvider
	workspace *workspace.Workspace
}

func NewPlanner(workspace *workspace.Workspace) Planner {
	return Planner{
		llm:       llm.NewOpenRouterProvider(),
		workspace: workspace,
	}
}

func CreatePlanStep(stepId int, task string) PlanStep {
	return PlanStep{
		StepId:   stepId,
		StepTask: task,
	}
}

func (p *Planner) ClassifyIntent(prompt string, projectType *ProjectTypeClassification) (*Intent, error) {
	projectTypeStr, err := json.Marshal(projectType)
	p.llm.SetModel(llm.MINIMAX_M2_5, true)
	intentStr, err := p.llm.Complete(prompts.IntentClassifier, prompt, string(projectTypeStr))

	fmt.Println(intentStr)

	if err != nil {
		return nil, err
	}
	var intent Intent
	err = json.Unmarshal([]byte(intentStr), &intent)
	if err != nil {
		return nil, err
	}
	return &intent, nil
}

type RepoType string
type ArchitectureStyle string
type DeploymentModel string

const (
	// Architecture styles
	ArchitectureMonolith     ArchitectureStyle = "monolith"
	ArchitectureMicroservice ArchitectureStyle = "microservice"
	ArchitectureMonorepo     ArchitectureStyle = "monorepo"
	ArchitectureLibrary      ArchitectureStyle = "library"
	ArchitectureCLI          ArchitectureStyle = "cli"
	ArchitectureHybrid       ArchitectureStyle = "hybrid"
	ArchitectureUnknown      ArchitectureStyle = "unknown"

	// Deployment models
	DeploymentServer        DeploymentModel = "server"
	DeploymentServerless    DeploymentModel = "serverless"
	DeploymentStatic        DeploymentModel = "static"
	DeploymentContainerized DeploymentModel = "containerized"
	DeploymentLibrary       DeploymentModel = "library"
	DeploymentUnknown       DeploymentModel = "unknown"
)

type ProjectTypeClassification struct {
	PrimaryType        RepoType          `json:"primary_type"`
	SecondaryTypes     []RepoType        `json:"secondary_types"`
	LanguagesDetected  []string          `json:"languages_detected"`
	FrameworksDetected []string          `json:"frameworks_detected"`
	ArchitectureStyle  ArchitectureStyle `json:"architecture_style"`
	DeploymentModel    DeploymentModel   `json:"deployment_model"`
	Confidence         float64           `json:"confidence"`
	Reasoning          string            `json:"reasoning"`
}

func (p *Planner) StartConversation() (string, error) {
	p.llm.SetModel(llm.MINIMAX_M2_5, true)

	value, err := p.llm.Complete(prompts.MainSystemPrompt, "How does the notification system work in the project?")

	if err != nil {
		return "", err
	}

	return value, nil
}

func (p *Planner) ClassifyProjectType() (*ProjectTypeClassification, error) {
	snapshot, err := GenerateRepoSnapshot(p.workspace.RootPath)
	if err != nil {
		return nil, err
	}

	snapshotStr, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}
	str, err := p.llm.Complete(prompts.ProjectTypeClassifier, string(snapshotStr))

	var projectType ProjectTypeClassification
	err = json.Unmarshal([]byte(str), &projectType)

	if err != nil {
		return nil, err
	}

	return &projectType, nil
}

func (p *Planner) CreatePlan(prompt string, intent *Intent, workspace *workspace.Workspace) Plan {
	initialStep := 0
	steps := []PlanStep{}

	for i := 0; i < 10; i++ {
		steps = append(steps, CreatePlanStep(initialStep, fmt.Sprintf("Task: %d", initialStep)))
		initialStep++
	}

	return Plan{
		Steps: steps,
	}
}

type StepDecision int

const (
	Allowed StepDecision = iota
	Blocked
	NeedApproval
)

type StepValidationResult struct {
	Valid    bool
	Error    string
	Warning  string
	Decision StepDecision
}

func (p Planner) ValidatePlan(plan *Plan) []StepValidationResult {
	validationResult := []StepValidationResult{}

	for i, step := range plan.Steps {
		if i == 2 {
			validationResult = append(
				validationResult,
				StepValidationResult{
					Valid:    false,
					Error:    "Invariant violation",
					Decision: Blocked,
				},
			)
		}

		validationResult = append(validationResult, p.ValidateStep(&step))
	}
	return validationResult
}

func (p Planner) ValidateStep(step *PlanStep) StepValidationResult {
	return StepValidationResult{Valid: true}
}

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

type GrepMatch struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

func ParseGrepOutput(r io.Reader) (*[]GrepMatch, error) {
	scanner := bufio.NewScanner(r)
	results := make([]GrepMatch, 0, 128)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Split only on last colon (safer for Windows paths)
		idx := strings.LastIndex(line, ":")
		if idx == -1 {
			continue
		}

		path := line[:idx]
		lineStr := line[idx+1:]

		count, err := strconv.Atoi(lineStr)
		if err != nil {
			continue
		}

		if strings.Contains(path, ".ts") || strings.Contains(path, ".tsx") || strings.Contains(path, ".go") {
			results = append(results, GrepMatch{
				Path:  path,
				Count: count,
			})
		}

	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &results, nil
}

func MostCommonPaths(groups *[][]GrepMatch) []string {
	if groups == nil {
		return nil
	}

	frequency := make(map[string]int)

	for _, group := range *groups {
		// Deduplicate paths inside a single group
		seen := make(map[string]struct{})

		for _, match := range group {
			if _, exists := seen[match.Path]; exists {
				continue
			}
			seen[match.Path] = struct{}{}
		}

		// Increment frequency once per group
		for path := range seen {
			frequency[path]++
		}
	}

	// Convert to slice for sorting
	type pair struct {
		Path  string
		Score int
	}

	pairs := make([]pair, 0, len(frequency))
	for path, score := range frequency {
		pairs = append(pairs, pair{Path: path, Score: score})
	}

	// Sort descending by frequency
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Score == pairs[j].Score {
			return pairs[i].Path < pairs[j].Path // stable deterministic ordering
		}
		return pairs[i].Score > pairs[j].Score
	})

	// Extract only paths
	result := make([]string, len(pairs))
	for i, p := range pairs {
		result[i] = p.Path
	}

	return result
}

func (p *Planner) ResolveScope(searchIdentifiers []string) ([]string, error) {

	inputList := [][]GrepMatch{{}}
	for _, pattern := range searchIdentifiers {
		cmd := exec.Command("rg", "-ic", pattern, p.workspace.RootPath)
		stdout, _ := cmd.StdoutPipe()
		_ = cmd.Start()
		output, err := ParseGrepOutput(stdout)
		if err != nil {
			return nil, err
		}

		inputList = append(inputList, *output)
	}

	frequentPaths := MostCommonPaths(&inputList)

	fmt.Println(frequentPaths)
	return frequentPaths, nil
}

func (p *Planner) GenerateChanges(prompt string, intent *Intent, files []string) (string, error) {
	p.llm.SetModel(llm.MINIMAX_M2_5, true)
	fileContext := ""

	for _, filename := range files {
		fileContent, err := os.ReadFile(filename)

		if err != nil {
			return "", err
		}

		fileContext = fmt.Sprintf("%s\n%s\n-----------\n%s", fileContext, filename, fileContent)
	}

	result, err := p.llm.Complete(prompts.ImplementFeature, prompt, fileContext, fmt.Sprintf("Requires new files: %b", intent.RequiresNewFiles))

	if err != nil {
		return "", err
	}

	fmt.Println(result)

	return "", nil
}
