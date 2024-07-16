package state_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

type MockTable struct {
	ID   int
	Name string
}

type MockRole struct {
	ID   int
	Name string
}

type MockRpc struct {
	ID   int
	Name string
}

type MockStorage struct {
	ID   string
	Name string
}

func TestLocalState_AddTable(t *testing.T) {
	localState := &state.LocalState{}
	table := state.TableState{
		Table: objects.Table{Name: "test_table"},
	}

	localState.AddTable(table)
	assert.Len(t, localState.State.Tables, 1)
	assert.True(t, localState.NeedUpdate)
}

func TestLocalState_FindTable(t *testing.T) {
	localState := &state.LocalState{}
	table := state.TableState{
		Table: objects.Table{ID: 1, Name: "test_table"},
	}
	localState.AddTable(table)

	index, tableState, found := localState.FindTable(1)
	assert.True(t, found)
	assert.Equal(t, 0, index)
	assert.Equal(t, table, tableState)
}

func TestLocalState_UpdateTable(t *testing.T) {
	localState := &state.LocalState{}
	table := state.TableState{
		Table: objects.Table{ID: 1, Name: "test_table"},
	}
	localState.AddTable(table)

	newTable := state.TableState{
		Table: objects.Table{ID: 1, Name: "updated_table"},
	}
	localState.UpdateTable(0, newTable)
	assert.Equal(t, "updated_table", localState.State.Tables[0].Table.Name)
}

func TestLocalState_DeleteTable(t *testing.T) {
	localState := &state.LocalState{}
	table := state.TableState{
		Table: objects.Table{ID: 1, Name: "test_table"},
	}
	localState.AddTable(table)

	localState.DeleteTable(1)
	assert.Empty(t, localState.State.Tables)
}

func TestLocalState_AddRole(t *testing.T) {
	localState := &state.LocalState{}
	role := state.RoleState{
		Role: objects.Role{Name: "test_role"},
	}

	localState.AddRole(role)
	assert.Len(t, localState.State.Roles, 1)
	assert.True(t, localState.NeedUpdate)
}

func TestLocalState_FindRole(t *testing.T) {
	localState := &state.LocalState{}
	role := state.RoleState{
		Role: objects.Role{ID: 1, Name: "test_role"},
	}
	localState.AddRole(role)

	index, roleState, found := localState.FindRole(1)
	assert.True(t, found)
	assert.Equal(t, 0, index)
	assert.Equal(t, role, roleState)
}

func TestLocalState_UpdateRole(t *testing.T) {
	localState := &state.LocalState{}
	role := state.RoleState{
		Role: objects.Role{ID: 1, Name: "test_role"},
	}
	localState.AddRole(role)

	newRole := state.RoleState{
		Role: objects.Role{ID: 1, Name: "updated_role"},
	}
	localState.UpdateRole(0, newRole)
	assert.Equal(t, "updated_role", localState.State.Roles[0].Role.Name)
}

func TestLocalState_DeleteRole(t *testing.T) {
	localState := &state.LocalState{}
	role := state.RoleState{
		Role: objects.Role{ID: 1, Name: "test_role"},
	}
	localState.AddRole(role)

	localState.DeleteRole(1)
	assert.Empty(t, localState.State.Roles)
}

func TestLocalState_AddRpc(t *testing.T) {
	localState := &state.LocalState{}
	rpc := state.RpcState{
		Function: objects.Function{Name: "test_rpc"},
	}

	localState.AddRpc(rpc)
	assert.Len(t, localState.State.Rpc, 1)
	assert.True(t, localState.NeedUpdate)
}

func TestLocalState_FindRpc(t *testing.T) {
	localState := &state.LocalState{}
	rpc := state.RpcState{
		Function: objects.Function{ID: 1, Name: "test_rpc"},
	}
	localState.AddRpc(rpc)

	index, rpcState, found := localState.FindRpc(1)
	assert.True(t, found)
	assert.Equal(t, 0, index)
	assert.Equal(t, rpc, rpcState)
}

func TestLocalState_UpdateRpc(t *testing.T) {
	localState := &state.LocalState{}
	rpc := state.RpcState{
		Function: objects.Function{ID: 1, Name: "test_rpc"},
	}
	localState.AddRpc(rpc)

	newRpc := state.RpcState{
		Function: objects.Function{ID: 1, Name: "updated_rpc"},
	}
	localState.UpdateRpc(0, newRpc)
	assert.Equal(t, "updated_rpc", localState.State.Rpc[0].Function.Name)
}

func TestLocalState_DeleteRpc(t *testing.T) {
	localState := &state.LocalState{}
	rpc := state.RpcState{
		Function: objects.Function{ID: 1, Name: "test_rpc"},
	}
	localState.AddRpc(rpc)

	localState.DeleteRpc(1)
	assert.Empty(t, localState.State.Rpc)
}

func TestLocalState_AddStorage(t *testing.T) {
	localState := &state.LocalState{}
	storage := state.StorageState{
		Storage: objects.Bucket{Name: "test_storage"},
	}

	localState.AddStorage(storage)
	assert.Len(t, localState.State.Storage, 1)
	assert.True(t, localState.NeedUpdate)
}

func TestLocalState_FindStorage(t *testing.T) {
	localState := &state.LocalState{}
	storage := state.StorageState{
		Storage: objects.Bucket{ID: "1", Name: "test_storage"},
	}
	localState.AddStorage(storage)

	index, storageState, found := localState.FindStorage("1")
	assert.True(t, found)
	assert.Equal(t, 0, index)
	assert.Equal(t, storage, storageState)
}

func TestLocalState_FindStorageByPermissionName(t *testing.T) {
	localState := &state.LocalState{}
	storage := state.StorageState{
		Storage: objects.Bucket{Name: "test_storage"},
	}
	localState.AddStorage(storage)

	index, storageState, found := localState.FindStorageByPermissionName("storagetest_storage")
	assert.True(t, found)
	assert.Equal(t, 0, index)
	assert.Equal(t, storage, storageState)
}

func TestLocalState_FindStorageByName(t *testing.T) {
	localState := &state.LocalState{}
	storage := state.StorageState{
		Storage: objects.Bucket{Name: "test_storage"},
	}
	localState.AddStorage(storage)

	index, storageState, found := localState.FindStorageByName("test_storage")
	assert.True(t, found)
	assert.Equal(t, 0, index)
	assert.Equal(t, storage, storageState)
}

func TestLocalState_UpdateStorage(t *testing.T) {
	localState := &state.LocalState{}
	storage := state.StorageState{
		Storage: objects.Bucket{ID: "1", Name: "test_storage"},
	}
	localState.AddStorage(storage)

	newStorage := state.StorageState{
		Storage: objects.Bucket{ID: "1", Name: "updated_storage"},
	}
	localState.UpdateStorage(0, newStorage)
	assert.Equal(t, "updated_storage", localState.State.Storage[0].Storage.Name)
}

func TestLocalState_DeleteStorage(t *testing.T) {
	localState := &state.LocalState{}
	storage := state.StorageState{
		Storage: objects.Bucket{ID: "1", Name: "test_storage"},
	}
	localState.AddStorage(storage)

	localState.DeleteStorage("1")
	assert.Empty(t, localState.State.Storage)
}

func TestLocalState_Persist(t *testing.T) {
	localState := &state.LocalState{}
	localState.AddTable(state.TableState{
		Table: objects.Table{Name: "test_table"},
	})

	err := localState.Persist()
	assert.NoError(t, err)
}

func TestSave(t *testing.T) {
	stateData := &state.State{
		Tables: []state.TableState{
			{Table: objects.Table{Name: "test_table"}},
		},
	}
	err := state.Save(stateData)
	assert.NoError(t, err)
}

func TestGetStateFilePath(t *testing.T) {
	path, err := state.GetStateFilePath()
	assert.NoError(t, err)
	assert.NotEmpty(t, path)
}

func TestCreateTmpState(t *testing.T) {
	stateFile := filepath.Join(os.TempDir(), "state_test")
	defer os.Remove(stateFile)

	err := os.WriteFile(stateFile, []byte("test"), 0644)
	assert.NoError(t, err)

	tmpFile := state.CreateTmpState(stateFile)
	assert.NotEmpty(t, tmpFile)
	defer os.Remove(tmpFile)
}

func TestRestoreFromTmp(t *testing.T) {
	stateFile := filepath.Join(os.TempDir(), "state_test")
	tmpFile := stateFile + ".tmp"
	defer os.Remove(stateFile)
	defer os.Remove(tmpFile)

	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	assert.NoError(t, err)

	state.RestoreFromTmp(tmpFile)
}

func TestLoad(t *testing.T) {
	_, err := state.Load()
	assert.NoError(t, err)
}
