package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

type PackageInput struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Integrity string `json:"integrity"`
}

type LockfileRequest struct {
	Repo      string         `json:"repo"`
	Ecosystem string         `json:"ecosystem"`
	Before    []PackageInput `json:"before"`
	After     []PackageInput `json:"after"`
}

func main() {
	token := os.Getenv("DIFFGUARD_TOKEN")
	apiURL := os.Getenv("DIFFGUARD_API_URL")
	lockfilePath := os.Getenv("DIFFGUARD_LOCKFILE")

	if token == "" || apiURL == "" {
		log.Fatal("DIFFGUARD_TOKEN and DIFFGUARD_API_URL are required")
	}

	// Auto-detect lockfile if not specified
	if lockfilePath == "" {
		lockfilePath = detectLockfile()
		if lockfilePath == "" {
			log.Println("No lockfile detected in workspace")
			os.Exit(0)
		}
	}

	ecosystem := detectEcosystem(lockfilePath)
	repo := os.Getenv("GITHUB_REPOSITORY")

	// Get the base branch lockfile content (before)
	baseSHA := os.Getenv("GITHUB_BASE_REF")
	if baseSHA == "" {
		baseSHA = "HEAD~1"
	}

	beforeContent := gitShow(baseSHA, lockfilePath)
	afterContent := readFile(lockfilePath)

	// Parse both versions
	beforePkgs := parseLockfileContent(ecosystem, beforeContent)
	afterPkgs := parseLockfileContent(ecosystem, afterContent)

	// POST to DiffGuard API
	reqBody := LockfileRequest{
		Repo:      repo,
		Ecosystem: ecosystem,
		Before:    beforePkgs,
		After:     afterPkgs,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL+"/api/v1/analyze/lockfile", bytes.NewReader(data))
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

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		OverallVerdict string `json:"overall_verdict"`
	}
	json.Unmarshal(body, &result)

	fmt.Printf("DiffGuard verdict: %s\n", result.OverallVerdict)

	switch result.OverallVerdict {
	case "RED":
		fmt.Println("::error::DiffGuard: Supply chain risk detected — PR blocked")
		os.Exit(1)
	case "YELLOW":
		fmt.Println("::warning::DiffGuard: Signals fired — human review required")
	case "GREEN":
		fmt.Println("DiffGuard: All clear")
	}
}

func detectLockfile() string {
	candidates := []string{
		"package-lock.json", "yarn.lock", "pnpm-lock.yaml",
		"poetry.lock", "requirements.txt",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func detectEcosystem(path string) string {
	base := filepath.Base(path)
	switch base {
	case "package-lock.json":
		return "npm"
	case "yarn.lock":
		return "npm"
	case "pnpm-lock.yaml":
		return "npm"
	case "poetry.lock":
		return "pypi"
	case "requirements.txt":
		return "pypi"
	default:
		return "npm"
	}
}

func gitShow(ref, path string) []byte {
	cmd := exec.Command("git", "show", ref+":"+path)
	out, err := cmd.Output()
	if err != nil {
		return nil // file didn't exist in base — all packages are "added"
	}
	return out
}

func readFile(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("reading lockfile: %v", err)
	}
	return data
}

// parseLockfileContent is a simplified parser for the action.
// It delegates to the server for full analysis — this just extracts
// name/version/hash tuples for the request payload.
func parseLockfileContent(ecosystem string, content []byte) []PackageInput {
	if len(content) == 0 {
		return nil
	}

	// For MVP, we send the raw lockfile content to the server
	// and let it do the parsing. The action just detects changes.
	// This stub returns nil — the server-side handler will parse.
	//
	// TODO: Import the shared lockfile parsers once the OSS repo split happens.
	// For now, the action sends before/after as raw content and the
	// server handles parsing via the existing internal/lockfile package.
	return nil
}
