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
	"strings"

	"github.com/diffguard/diffguard/internal/lockfile"
)

type PackageInput struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Integrity string `json:"integrity"`
}

type ResolutionRequest struct {
	Repo      string         `json:"repo"`
	Ecosystem string         `json:"ecosystem"`
	Packages  []PackageInput `json:"packages"`
}

func main() {
	token := os.Getenv("DIFFGUARD_TOKEN")
	apiURL := os.Getenv("DIFFGUARD_API_URL")
	lockfilePath := os.Getenv("DIFFGUARD_LOCKFILE")

	if token == "" || apiURL == "" {
		log.Fatal("DIFFGUARD_TOKEN and DIFFGUARD_API_URL are required")
	}

	if lockfilePath == "" {
		log.Fatal("DIFFGUARD_LOCKFILE is required")
	}

	// Check lockfile exists
	if _, err := os.Stat(lockfilePath); os.IsNotExist(err) {
		fmt.Println("::error::DiffGuard: No lockfile found at", lockfilePath)
		fmt.Println("::error::DiffGuard: CRITICAL — no lockfile present. Commit a lockfile to enable dependency monitoring.")
		os.Exit(1)
	}

	f, err := os.Open(lockfilePath)
	if err != nil {
		log.Fatalf("opening lockfile: %v", err)
	}
	defer f.Close()

	ecosystem := detectEcosystem(lockfilePath)
	pkgs, err := parseLockfile(ecosystem, f)
	if err != nil {
		log.Fatalf("parsing lockfile: %v", err)
	}

	repo := os.Getenv("GITHUB_REPOSITORY")

	// Convert to API format
	inputs := make([]PackageInput, len(pkgs))
	for i, p := range pkgs {
		inputs[i] = PackageInput{
			Name:      p.Name,
			Version:   p.Version,
			Integrity: p.Integrity,
		}
	}

	reqBody := ResolutionRequest{
		Repo:      repo,
		Ecosystem: ecosystem,
		Packages:  inputs,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL+"/api/v1/analyze/resolution", bytes.NewReader(data))
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

	fmt.Printf("DiffGuard Passive Monitor verdict: %s\n", result.OverallVerdict)

	if result.OverallVerdict == "RED" {
		fmt.Println("::error::DiffGuard: Supply chain risk detected in resolved dependencies")
		os.Exit(1)
	}
}

func detectEcosystem(path string) string {
	base := filepath.Base(path)
	switch base {
	case "package-lock.json", "yarn.lock", "pnpm-lock.yaml":
		return "npm"
	case "poetry.lock", "requirements.txt":
		return "pypi"
	default:
		return "npm"
	}
}

func parseLockfile(ecosystem string, f *os.File) ([]lockfile.Package, error) {
	base := filepath.Base(f.Name())
	switch {
	case base == "package-lock.json":
		return lockfile.ParseNPM(f)
	case base == "yarn.lock":
		return lockfile.ParseYarn(f)
	case base == "pnpm-lock.yaml":
		return lockfile.ParsePNPM(f)
	case base == "poetry.lock":
		return lockfile.ParsePoetry(f)
	case strings.HasSuffix(base, "requirements.txt"):
		return lockfile.ParsePip(f)
	default:
		return nil, fmt.Errorf("unsupported lockfile: %s", base)
	}
}
