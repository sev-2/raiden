package acl

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/require"
)

func prepareState(t *testing.T, st *state.State) {
	t.Helper()

	relativeDir := filepath.Join("testdata", "tmp", t.Name())

	originalDir := state.StateFileDir
	originalFile := state.StateFileName

	state.StateFileDir = relativeDir
	state.StateFileName = fmt.Sprintf("state_%s", t.Name())

	curDir, err := utils.GetCurrentDirectory()
	require.NoError(t, err)

	statePath := filepath.Join(curDir, relativeDir)
	require.NoError(t, os.MkdirAll(statePath, 0o755))

	t.Cleanup(func() {
		state.StateFileDir = originalDir
		state.StateFileName = originalFile
		require.NoError(t, os.RemoveAll(statePath))
	})

	require.NoError(t, state.Save(st))
}
