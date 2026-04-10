package lockfile_test

import (
	"os"
	"testing"

	"github.com/diffguard/actions/shared/lockfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePip(t *testing.T) {
	f, err := os.Open("testdata/requirements.txt")
	require.NoError(t, err)
	defer f.Close()

	pkgs, err := lockfile.ParsePip(f)
	require.NoError(t, err)
	require.Len(t, pkgs, 3)

	assert.Equal(t, "certifi", pkgs[0].Name)
	assert.Equal(t, "2023.7.22", pkgs[0].Version)
	assert.Equal(t, "sha256-92d6037539857d8206b8f6ae472e8b77db8058fec5937a1ef3f54304089edbb9", pkgs[0].Integrity)
	assert.Equal(t, "pypi", pkgs[0].Ecosystem)

	assert.Equal(t, "requests", pkgs[1].Name)
	assert.Equal(t, "2.31.0", pkgs[1].Version)
	// First hash encountered wins
	assert.Equal(t, "sha256-58cd2187423839a866e8fdeee2b4cf294f9e93ccd3bc15f9da6d24a5cb86765c", pkgs[1].Integrity)

	assert.Equal(t, "urllib3", pkgs[2].Name)
	assert.Equal(t, "2.0.7", pkgs[2].Version)
}
