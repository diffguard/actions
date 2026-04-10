package lockfile_test

import (
	"os"
	"testing"

	"github.com/diffguard/actions/shared/lockfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseYarn(t *testing.T) {
	f, err := os.Open("testdata/yarn.lock")
	require.NoError(t, err)
	defer f.Close()

	pkgs, err := lockfile.ParseYarn(f)
	require.NoError(t, err)
	require.Len(t, pkgs, 2)

	// Results sorted by name
	assert.Equal(t, "chalk", pkgs[0].Name)
	assert.Equal(t, "4.1.2", pkgs[0].Version)
	assert.Equal(t, "sha512-oKnbhFyRIXpUuez8iBMmyEa4nbj4IOQyuhc/wy9kY7/WVPcwIO9VA668Pu8RkO7+0G76SLROeyw9CpQ061i4A==", pkgs[0].Integrity)
	assert.Equal(t, "npm", pkgs[0].Ecosystem)

	assert.Equal(t, "lodash", pkgs[1].Name)
	assert.Equal(t, "4.17.21", pkgs[1].Version)
}
