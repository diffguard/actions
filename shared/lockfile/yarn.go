package lockfile

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

// ParseYarn parses a yarn.lock (v1 classic format) and returns all resolved packages.
// Yarn Berry (v2+) uses YAML — not supported in MVP.
func ParseYarn(r io.Reader) ([]Package, error) {
	scanner := bufio.NewScanner(r)

	// Deduplicate: yarn.lock may list the same resolved version under multiple specifier aliases.
	seen := map[string]struct{}{}
	var pkgs []Package

	var currentName string
	var currentVersion, currentIntegrity string

	flush := func() {
		if currentName == "" || currentVersion == "" {
			return
		}
		key := currentName + "@" + currentVersion
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			pkgs = append(pkgs, Package{
				Ecosystem: "npm",
				Name:      currentName,
				Version:   currentVersion,
				Integrity: currentIntegrity,
			})
		}
		currentName = ""
		currentVersion = ""
		currentIntegrity = ""
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and blank lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			flush()
			continue
		}

		// Header line: `lodash@^4.17.21:` or `chalk@^4.0.0, chalk@^4.1.0:`
		if !strings.HasPrefix(line, " ") && strings.HasSuffix(strings.TrimSpace(line), ":") {
			flush()
			// Take the first specifier's package name (before the @version specifier)
			spec := strings.TrimSuffix(strings.TrimSpace(line), ":")
			firstSpec := strings.Split(spec, ",")[0]
			firstSpec = strings.TrimSpace(firstSpec)
			// Package name is everything before the last @ that starts a version specifier
			// e.g. "lodash@^4.17.21" → "lodash"; "@babel/core@^7.0.0" → "@babel/core"
			atIdx := strings.LastIndex(firstSpec, "@")
			if atIdx > 0 {
				currentName = firstSpec[:atIdx]
			}
			continue
		}

		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "version ") {
			currentVersion = strings.Trim(strings.TrimPrefix(trimmed, "version "), "\"")
		}

		if strings.HasPrefix(trimmed, "integrity ") {
			currentIntegrity = strings.TrimPrefix(trimmed, "integrity ")
		}
	}

	flush()

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning yarn.lock: %w", err)
	}

	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].Name < pkgs[j].Name })
	return pkgs, nil
}
