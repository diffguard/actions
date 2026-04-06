package lockfile_test

import (
	"os"
	"testing"

	"github.com/diffguard/diffguard/actions/shared/lockfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePoetry(t *testing.T) {
	f, err := os.Open("testdata/poetry.lock")
	require.NoError(t, err)
	defer f.Close()

	pkgs, err := lockfile.ParsePoetry(f)
	require.NoError(t, err)
	require.Len(t, pkgs, 2)

	assert.Equal(t, "requests", pkgs[0].Name)
	assert.Equal(t, "2.31.0", pkgs[0].Version)
	// Uses the first file's hash
	assert.Equal(t, "sha256-58cd2187423839a866e8fdeee2b4cf294f9e93ccd3bc15f9da6d24a5cb86765c", pkgs[0].Integrity)
	assert.Equal(t, "pypi", pkgs[0].Ecosystem)

	assert.Equal(t, "urllib3", pkgs[1].Name)
	assert.Equal(t, "2.0.7", pkgs[1].Version)
}
