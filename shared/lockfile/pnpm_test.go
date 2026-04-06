package lockfile_test

import (
	"os"
	"testing"

	"github.com/diffguard/diffguard/actions/shared/lockfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePNPM(t *testing.T) {
	f, err := os.Open("testdata/pnpm-lock.yaml")
	require.NoError(t, err)
	defer f.Close()

	pkgs, err := lockfile.ParsePNPM(f)
	require.NoError(t, err)
	require.Len(t, pkgs, 2)

	assert.Equal(t, "express", pkgs[0].Name)
	assert.Equal(t, "4.18.2", pkgs[0].Version)
	assert.Equal(t, "sha512-o45mr9nAtSjQMPfBo5x1f5e4a7j6Sx+YXiEjO6NkL/u9DMj3FmLF/dBgIelRgSn7S6BVsC0f0e8Gw9vlMMoAA==", pkgs[0].Integrity)
	assert.Equal(t, "npm", pkgs[0].Ecosystem)

	assert.Equal(t, "lodash", pkgs[1].Name)
	assert.Equal(t, "4.17.21", pkgs[1].Version)
}
