package state

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"time"

	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
)

var StateFileDir = "build"
var StateFileName = "state"

type State struct {
	Tables   []TableState
	Roles    []RoleState
	RpcState []RpcState
}

type TableState struct {
	Table       supabase.Table
	ModelPath   string
	ModelStruct string
	LastUpdate  time.Time
}

type RoleState struct {
	Role       supabase.Role
	RolePath   string
	RoleStruct string
	LastUpdate time.Time
}

type RpcState struct {
	Function   supabase.Function
	RpcPath    string
	RpcStruct  string
	LastUpdate time.Time
}

func Save(state *State) error {
	curDir, err := utils.GetCurrentDirectory()
	if err != nil {
		return err
	}

	statePath := filepath.Join(curDir, StateFileDir)
	if !utils.IsFolderExists(statePath) {
		if err := utils.CreateFolder(statePath); err != nil {
			return err
		}
	}

	filePath := filepath.Join(statePath, StateFileName)
	file, err := createOrLoadFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	logger.Debug("Save state file to : ", filePath)

	encoder := gob.NewEncoder(file)
	return encoder.Encode(state)
}

func Load() (*State, error) {
	curDir, err := utils.GetCurrentDirectory()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(curDir, StateFileDir, StateFileName)
	if !utils.IsFileExists(filePath) {
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	state := &State{}
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(state); err != nil {
		return nil, err
	}
	return state, nil
}

func createOrLoadFile(filePath string) (file *os.File, err error) {
	if utils.IsFileExists(filePath) {
		return os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	}
	return os.Create(filePath)
}
