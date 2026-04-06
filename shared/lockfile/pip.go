package lockfile

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

// ParsePip parses a pip-tools compiled requirements.txt with --hash entries.
// Lines ending with `\` are continuation lines. Bare pip requirements without
// hashes are included with an empty Integrity field.
func ParsePip(r io.Reader) ([]Package, error) {
	scanner := bufio.NewScanner(r)

	// Join continuation lines first
	var lines []string
	var current strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasSuffix(line, "\\") {
			current.WriteString(strings.TrimSuffix(line, "\\"))
			current.WriteString(" ")
		} else {
			current.WriteString(line)
			lines = append(lines, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning requirements.txt: %w", err)
	}

	var pkgs []Package

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}

		// Extract package==version
		// Format: "name==version --hash=sha256:hex [--hash=sha256:hex ...]"
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		spec := parts[0]
		eqIdx := strings.Index(spec, "==")
		if eqIdx == -1 {
			continue
		}
		name := spec[:eqIdx]
		version := spec[eqIdx+2:]

		// Find the first --hash= value
		var firstHash string
		for _, part := range parts[1:] {
			if strings.HasPrefix(part, "--hash=sha256:") {
				h := strings.TrimPrefix(part, "--hash=sha256:")
				firstHash = "sha256-" + h
				break
			}
		}

		pkgs = append(pkgs, Package{
			Ecosystem: "pypi",
			Name:      name,
			Version:   version,
			Integrity: firstHash,
		})
	}

	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].Name < pkgs[j].Name })
	return pkgs, nil
}
