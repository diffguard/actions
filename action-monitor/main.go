package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ActionRef represents a single `uses:` reference from a workflow file.
type ActionRef struct {
	Owner    string `json:"owner"`
	Repo     string `json:"repo"`
	Ref      string `json:"ref"`      // tag, branch, or SHA
	RefType  string `json:"ref_type"` // "sha", "tag", "branch", "local"
	Workflow string `json:"workflow"` // source workflow file
}

type AnalyzeActionsRequest struct {
	Repo    string      `json:"repo"`
	Actions []ActionRef `json:"actions"`
}

var shaRegex = regexp.MustCompile(`^[0-9a-f]{40}$`)

func main() {
	token := os.Getenv("DIFFGUARD_TOKEN")
	apiURL := os.Getenv("DIFFGUARD_API_URL")

	if token == "" || apiURL == "" {
		log.Fatal("DIFFGUARD_TOKEN and DIFFGUARD_API_URL are required")
	}

	repo := os.Getenv("GITHUB_REPOSITORY")

	// Find all workflow files
	workflowDir := ".github/workflows"
	entries, err := os.ReadDir(workflowDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No .github/workflows directory found")
			os.Exit(0)
		}
		log.Fatalf("reading workflows dir: %v", err)
	}

	var refs []ActionRef
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			continue
		}

		path := filepath.Join(workflowDir, name)
		parsed, err := parseWorkflow(path)
		if err != nil {
			log.Printf("warning: failed to parse %s: %v", path, err)
			continue
		}
		for i := range parsed {
			parsed[i].Workflow = name
		}
		refs = append(refs, parsed...)
	}

	if len(refs) == 0 {
		fmt.Println("No action references found in workflows")
		os.Exit(0)
	}

	fmt.Printf("Found %d action references across %d workflow files\n", len(refs), len(entries))

	// POST to DiffGuard API
	reqBody := AnalyzeActionsRequest{
		Repo:    repo,
		Actions: refs,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL+"/api/v1/analyze/actions", bytes.NewReader(data))
	if err != nil {
		log.Fatalf("creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("calling DiffGuard API: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusNotImplemented {
		fmt.Println("Action analysis not yet implemented on server — skipping")
		os.Exit(0)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("API returned %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("DiffGuard Action Monitor: %s\n", string(body))
}

// Workflow represents the minimal structure of a GitHub Actions workflow YAML.
type Workflow struct {
	Jobs map[string]Job `yaml:"jobs"`
}

type Job struct {
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Uses string `yaml:"uses"`
}

func parseWorkflow(path string) ([]ActionRef, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var wf Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, err
	}

	var refs []ActionRef
	seen := make(map[string]bool)

	for _, job := range wf.Jobs {
		for _, step := range job.Steps {
			if step.Uses == "" {
				continue
			}
			if seen[step.Uses] {
				continue
			}
			seen[step.Uses] = true

			ref := classifyRef(step.Uses)
			refs = append(refs, ref)
		}
	}
	return refs, nil
}

func classifyRef(uses string) ActionRef {
	// Local action: ./path
	if strings.HasPrefix(uses, "./") || strings.HasPrefix(uses, "../") {
		return ActionRef{RefType: "local", Ref: uses}
	}

	// Docker action: docker://...
	if strings.HasPrefix(uses, "docker://") {
		return ActionRef{RefType: "local", Ref: uses}
	}

	// Parse owner/repo@ref
	atIdx := strings.LastIndex(uses, "@")
	if atIdx == -1 {
		return ActionRef{RefType: "tag", Ref: uses}
	}

	ownerRepo := uses[:atIdx]
	ref := uses[atIdx+1:]

	parts := strings.SplitN(ownerRepo, "/", 2)
	owner := ""
	repo := ""
	if len(parts) == 2 {
		owner = parts[0]
		repo = parts[1]
	}

	refType := "tag"
	if shaRegex.MatchString(ref) {
		refType = "sha"
	} else if ref == "main" || ref == "master" || ref == "develop" {
		refType = "branch"
	}

	return ActionRef{
		Owner:   owner,
		Repo:    repo,
		Ref:     ref,
		RefType: refType,
	}
}
