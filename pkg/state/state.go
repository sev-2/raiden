package state

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type (
	State struct {
		Tables  []TableState
		Roles   []RoleState
		Rpc     []RpcState
		Storage []StorageState
	}

	TableState struct {
		Table       objects.Table
		Relation    []Relation
		ModelPath   string
		ModelStruct string
		LastUpdate  time.Time
		Policies    []objects.Policy
	}

	RoleState struct {
		Role       objects.Role
		RolePath   string
		RoleStruct string
		IsNative   bool
		LastUpdate time.Time
	}

	RpcState struct {
		Function   objects.Function
		RpcPath    string
		RpcStruct  string
		LastUpdate time.Time
	}

	StorageState struct {
		Storage    objects.Storage
		LastUpdate time.Time
	}

	Relation struct {
		Table        string
		Type         string
		RelationType raiden.RelationType
		PrimaryKey   string
		ForeignKey   string
		Tag          string
		*JoinRelation
	}

	JoinRelation struct {
		SourcePrimaryKey      string
		JoinsSourceForeignKey string

		TargetPrimaryKey     string
		JoinTargetForeignKey string

		Through string
	}
)

var (
	StateFileDir  = "build"
	StateFileName = "state"
)

func Save(state *State) error {
	filePath, err := GetStateFilePath()
	if err != nil {
		return err
	}

	var tmpFilePath string
	if exist := utils.IsFileExists(filePath); exist {
		tmpFilePath = CreateTmpState(filePath)
	}

	file, err := createOrLoadFile(filePath)
	if err != nil {
		return err
	}

	defer file.Close()

	logger.Debug("State : Generate local state to : ", filePath)
	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(state); err != nil {
		RestoreFromTmp(tmpFilePath)
		return err
	}

	utils.DeleteFile(tmpFilePath)
	return nil
}

func GetStateFilePath() (path string, err error) {
	curDir, err := utils.GetCurrentDirectory()
	if err != nil {
		return path, err
	}

	statePath := filepath.Join(curDir, StateFileDir)
	if !utils.IsFolderExists(statePath) {
		if err := utils.CreateFolder(statePath); err != nil {
			return path, err
		}
	}

	return filepath.Join(statePath, StateFileName), nil
}

func CreateTmpState(stateFile string) string {
	filePathTmp := stateFile + ".tmp"
	if exist := utils.IsFileExists(stateFile); exist {
		utils.CopyFile(stateFile, filePathTmp)
	}

	return filePathTmp
}

func RestoreFromTmp(tmpFile string) {
	if utils.IsFileExists(tmpFile) {
		filePath := strings.TrimSuffix(tmpFile, ".tmp")
		utils.CopyFile(filePath, filePath)
		return
	}

	logger.Error("file is not exist : ", tmpFile)
}

func Load() (*State, error) {
	filePath, err := GetStateFilePath()
	if err != nil {
		return nil, err
	}

	if !utils.IsFileExists(filePath) {
		// save empty sta
		Save(&State{})
		return nil, nil
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
