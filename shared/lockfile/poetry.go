package lockfile

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

// ParsePoetry parses a poetry.lock (TOML-like format) and returns all resolved packages.
// We do a line-by-line parse to avoid a TOML dependency — poetry.lock format is stable.
func ParsePoetry(r io.Reader) ([]Package, error) {
	scanner := bufio.NewScanner(r)

	var pkgs []Package
	var currentName, currentVersion string
	var firstHash string
	inPackage := false

	flush := func() {
		if currentName != "" && currentVersion != "" {
			pkgs = append(pkgs, Package{
				Ecosystem: "pypi",
				Name:      currentName,
				Version:   currentVersion,
				Integrity: firstHash,
			})
		}
		currentName = ""
		currentVersion = ""
		firstHash = ""
		inPackage = false
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "[[package]]" {
			flush()
			inPackage = true
			continue
		}

		if !inPackage {
			continue
		}

		if strings.HasPrefix(trimmed, "name = ") {
			currentName = strings.Trim(strings.TrimPrefix(trimmed, "name = "), "\"")
		} else if strings.HasPrefix(trimmed, "version = ") {
			currentVersion = strings.Trim(strings.TrimPrefix(trimmed, "version = "), "\"")
		} else if firstHash == "" && strings.Contains(trimmed, "hash = \"sha256:") {
			// Extract sha256 from: {file = "...", hash = "sha256:abc123"}
			idx := strings.Index(trimmed, "hash = \"sha256:")
			if idx != -1 {
				rest := trimmed[idx+len("hash = \"sha256:"):]
				endIdx := strings.Index(rest, "\"")
				if endIdx != -1 {
					firstHash = "sha256-" + rest[:endIdx]
				}
			}
		}
	}

	flush()

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning poetry.lock: %w", err)
	}

	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].Name < pkgs[j].Name })
	return pkgs, nil
}
