package state

import (
	"encoding/gob"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type (
	State struct {
		Tables   []TableState
		Roles    []RoleState
		RpcState []RpcState
	}

	TableState struct {
		Table       objects.Table
		Relation    []generator.Relation
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
)

var (
	StateFileDir  = "build"
	StateFileName = "state"
)

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
	filePathTmp := filePath + ".tmp"
	if exist := utils.IsFileExists(filePath); exist {
		if err := utils.CopyFile(filePath, filePathTmp); err != nil {
			return err
		}
	}

	file, err := createOrLoadFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	logger.Debug("Save state file to : ", filePath)

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(state); err != nil {
		utils.DeleteFile(filePath)
		utils.CopyFile(filePathTmp, filePath)
		return err
	}

	utils.DeleteFile(filePathTmp)
	return nil
}

func Load() (*State, error) {
	curDir, err := utils.GetCurrentDirectory()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(curDir, StateFileDir, StateFileName)
	if !utils.IsFileExists(filePath) {
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

func loadFiles(paths []string) (mainFset *token.FileSet, astFiles []*ast.File, err error) {
	type loadResult struct {
		Ast  *ast.File
		Err  error
		Fset *token.FileSet
	}

	loadChan := make(chan *loadResult)

	wg := sync.WaitGroup{}
	wg.Add(len(paths))

	go func() {
		wg.Wait()
		close(loadChan)
	}()

	for _, path := range paths {
		go func(w *sync.WaitGroup, p string, lChan chan *loadResult) {
			defer w.Done()
			fset := token.NewFileSet()

			logger.Debug("Load file : ", p)
			file, err := parser.ParseFile(fset, p, nil, parser.ParseComments)
			if err != nil {
				lChan <- &loadResult{Err: err}
			} else {
				lChan <- &loadResult{Ast: file, Fset: fset, Err: nil}
			}
		}(&wg, path, loadChan)
	}

	for rs := range loadChan {
		if rs.Err != nil {
			if strings.Contains(rs.Err.Error(), "no such file or directory") {
				continue
			}
			return nil, nil, rs.Err
		}
		if mainFset == nil {
			mainFset = rs.Fset
		}

		astFiles = append(astFiles, rs.Ast)
	}
	return
}
