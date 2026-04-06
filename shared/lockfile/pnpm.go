package lockfile

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// pnpmLockV6 covers pnpm lockfile v6 (pnpm 8.x) "snapshots" key.
// Earlier versions use "packages" key with slightly different shape.
type pnpmLockV6 struct {
	LockfileVersion string                   `yaml:"lockfileVersion"`
	Snapshots       map[string]pnpmSnapshot  `yaml:"snapshots"`
	Packages        map[string]pnpmPackageV5 `yaml:"packages"` // v5 fallback
}

type pnpmSnapshot struct {
	Resolution pnpmResolution `yaml:"resolution"`
}

type pnpmResolution struct {
	Integrity string `yaml:"integrity"`
}

type pnpmPackageV5 struct {
	Resolution pnpmResolution `yaml:"resolution"`
}

// ParsePNPM parses a pnpm-lock.yaml (v5 or v6) and returns all resolved packages.
func ParsePNPM(r io.Reader) ([]Package, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading pnpm lockfile: %w", err)
	}

	var lf pnpmLockV6
	if err := yaml.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("parsing pnpm lockfile: %w", err)
	}

	var pkgs []Package

	parseKey := func(key, integrity string) {
		// Key format: "name@version" or "@scope/name@version"
		atIdx := strings.LastIndex(key, "@")
		if atIdx <= 0 {
			return
		}
		name := key[:atIdx]
		version := key[atIdx+1:]
		// Strip any trailing peer dep suffix e.g. "1.2.3(react@18.0.0)"
		if parenIdx := strings.Index(version, "("); parenIdx != -1 {
			version = version[:parenIdx]
		}
		pkgs = append(pkgs, Package{
			Ecosystem: "npm",
			Name:      name,
			Version:   version,
			Integrity: integrity,
		})
	}

	if len(lf.Snapshots) > 0 {
		for key, snap := range lf.Snapshots {
			parseKey(key, snap.Resolution.Integrity)
		}
	} else {
		for key, pkg := range lf.Packages {
			parseKey(key, pkg.Resolution.Integrity)
		}
	}

	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].Name < pkgs[j].Name })
	return pkgs, nil
}
