package state

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

var StateLogger = logger.HcLog().Named("supabase.state")

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
		Storage       objects.Bucket
		StoragePath   string
		StorageStruct string
		LastUpdate    time.Time
		Policies      []objects.Policy
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

	LocalState struct {
		State      State
		NeedUpdate bool
		Mutex      sync.RWMutex
	}

	ExtractedPolicies struct {
		Existing []objects.Policy
		New      []objects.Policy
		Delete   []objects.Policy
	}
)

var (
	StateFileDir  = "build"
	StateFileName = "state"
)

func (s *LocalState) AddTable(table TableState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.State.Tables = append(s.State.Tables, table)
	s.NeedUpdate = true
}

func (s *LocalState) FindTable(tableId int) (index int, tableState TableState, found bool) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	found = false

	for i := range s.State.Tables {
		t := s.State.Tables[i]

		if t.Table.ID == tableId {
			found = true
			tableState = t
			index = i
			return
		}
	}
	return
}

func (s *LocalState) UpdateTable(index int, state TableState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.State.Tables[index] = state
	s.NeedUpdate = true
}

func (s *LocalState) DeleteTable(tableId int) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	index := -1
	for i := range s.State.Tables {
		t := s.State.Tables[i]
		if t.Table.ID == tableId {
			index = i
			break
		}
	}

	if index == -1 {
		return
	}
	s.State.Tables = append(s.State.Tables[:index], s.State.Tables[index+1:]...)
	s.NeedUpdate = true
}

func (s *LocalState) AddRole(role RoleState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.State.Roles = append(s.State.Roles, role)
	s.NeedUpdate = true
}

func (s *LocalState) FindRole(roleId int) (index int, roleState RoleState, found bool) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	found = false

	for i := range s.State.Roles {
		r := s.State.Roles[i]

		if r.Role.ID == roleId {
			found = true
			roleState = r
			index = i
			return
		}
	}
	return
}

func (s *LocalState) UpdateRole(index int, state RoleState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.State.Roles[index] = state
	s.NeedUpdate = true
}

func (s *LocalState) DeleteRole(roleId int) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	index := -1
	for i := range s.State.Roles {
		r := s.State.Roles[i]

		if r.Role.ID == roleId {
			index = i
			break
		}
	}

	if index == -1 {
		return
	}
	s.State.Roles = append(s.State.Roles[:index], s.State.Roles[index+1:]...)
	s.NeedUpdate = true
}

func (s *LocalState) AddRpc(rpc RpcState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.State.Rpc = append(s.State.Rpc, rpc)
	s.NeedUpdate = true
}

func (s *LocalState) FindRpc(rpcId int) (index int, rpcState RpcState, found bool) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	found = false

	for i := range s.State.Rpc {
		r := s.State.Rpc[i]

		if r.Function.ID == rpcId {
			found = true
			rpcState = r
			index = i
			return
		}
	}
	return
}

func (s *LocalState) DeleteRpc(rpcId int) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	index := -1
	for i := range s.State.Rpc {
		r := s.State.Rpc[i]

		if r.Function.ID == rpcId {
			index = i
			break
		}
	}

	if index == -1 {
		return
	}
	s.State.Rpc = append(s.State.Rpc[:index], s.State.Rpc[index+1:]...)
	s.NeedUpdate = true
}

func (s *LocalState) UpdateRpc(index int, state RpcState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.State.Rpc[index] = state
	s.NeedUpdate = true
}

func (s *LocalState) AddStorage(storage StorageState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.State.Storage = append(s.State.Storage, storage)
	s.NeedUpdate = true
}

func (s *LocalState) FindStorage(storageId string) (index int, storageState StorageState, found bool) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	found = false

	for i := range s.State.Storage {
		r := s.State.Storage[i]

		if r.Storage.ID == storageId {
			found = true
			storageState = r
			index = i
			return
		}
	}
	return
}

func (s *LocalState) FindStorageByPermissionName(name string) (index int, storageState StorageState, found bool) {
	// find storage name
	splitName := strings.SplitN(name, supabase.RlsTypeStorage, 2)
	if len(splitName) != 2 {
		found = false
		return
	}

	storageName := strings.TrimLeft(strings.TrimRight(splitName[1], " "), " ")
	return s.FindStorageByName(storageName)
}

func (s *LocalState) FindStorageByName(name string) (index int, storageState StorageState, found bool) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	found = false

	for i := range s.State.Storage {
		r := s.State.Storage[i]
		if strings.EqualFold(r.Storage.Name, name) {
			found = true
			storageState = r
			index = i
			return
		}
	}
	return
}

func (s *LocalState) UpdateStorage(index int, state StorageState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.State.Storage[index] = state
	s.NeedUpdate = true
}

func (s *LocalState) DeleteStorage(storageId string) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	index := -1
	for i := range s.State.Storage {
		r := s.State.Storage[i]

		if r.Storage.ID == storageId {
			index = i
			break
		}
	}

	if index == -1 {
		return
	}
	s.State.Storage = append(s.State.Storage[:index], s.State.Storage[index+1:]...)
	s.NeedUpdate = true
}

func (s *LocalState) Persist() error {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	if s.NeedUpdate {
		if err := Save(&s.State); err != nil {
			return err
		}
		s.NeedUpdate = false
	}
	return nil
}

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

	StateLogger.Debug("generate local state", "path", filePath)
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

	StateLogger.Debug("file is not exist", "path", tmpFile)
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
