package lockfile

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

type npmLockfile struct {
	LockfileVersion int                    `json:"lockfileVersion"`
	Packages        map[string]npmPackage  `json:"packages"`
	Dependencies    map[string]npmDependency `json:"dependencies"`
}

type npmPackage struct {
	Version   string `json:"version"`
	Integrity string `json:"integrity"`
}

type npmDependency struct {
	Version   string `json:"version"`
	Resolved  string `json:"resolved"`
	Integrity string `json:"integrity"`
}

func ParseNPM(r io.Reader) ([]Package, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading npm lockfile: %w", err)
	}

	var lf npmLockfile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("parsing npm lockfile: %w", err)
	}

	var pkgs []Package

	if lf.LockfileVersion >= 2 && len(lf.Packages) > 0 {
		for key, p := range lf.Packages {
			if !strings.HasPrefix(key, "node_modules/") {
				continue
			}
			name := strings.TrimPrefix(key, "node_modules/")
			if p.Version == "" {
				continue
			}
			pkgs = append(pkgs, Package{
				Ecosystem: "npm",
				Name:      name,
				Version:   p.Version,
				Integrity: p.Integrity,
			})
		}
	} else {
		for name, dep := range lf.Dependencies {
			pkgs = append(pkgs, Package{
				Ecosystem: "npm",
				Name:      name,
				Version:   dep.Version,
				Integrity: dep.Integrity,
			})
		}
	}

	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].Name < pkgs[j].Name })
	return pkgs, nil
}
