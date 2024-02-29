package resource

import (
	"fmt"
	"strings"
	"sync"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type MigrateType string

const (
	MigrateTypeIgnore MigrateType = "ignore"
	MigrateTypeCreate MigrateType = "create"
	MigrateTypeUpdate MigrateType = "update"
	MigrateTypeDelete MigrateType = "delete"
)

type MigrateItem[T, D any] struct {
	Type           MigrateType
	NewData        T
	OldData        T
	MigrationItems D
}

type MigrateData struct {
	Tables   []MigrateItem[objects.Table, objects.UpdateTableParam]
	Roles    []MigrateItem[objects.Role, objects.UpdateRoleParam]
	Rpc      []MigrateItem[objects.Function, any]
	Policies []MigrateItem[objects.Policy, any]
}

type MigrateCreateFunc[T any] func(cfg *raiden.Config, param T) (response T, err error)
type MigrateUpdateFunc[T, D any] func(cfg *raiden.Config, param T, items D) (err error)
type MigrateDeleteFunc[T any] func(cfg *raiden.Config, param T) (err error)

type MigrateActionFunc[T, D any] struct {
	CreateFunc MigrateCreateFunc[T]
	UpdateFunc MigrateUpdateFunc[T, D]
	DeleteFunc MigrateDeleteFunc[T]
}
type MigrateFuncParam[T, D any] struct {
	Config      *raiden.Config
	StateChan   chan any
	Data        MigrateItem[T, D]
	ActionFuncs MigrateActionFunc[T, D]
}
type MigrateFunc[T, D any] func(param MigrateFuncParam[T, D]) error
type MigrateTableNode struct {
	Level         int
	MigratedItems []MigrateItem[objects.Table, objects.UpdateTableParam]
	Child         *MigrateTableNode
}

func MigrateResource(config *raiden.Config, importState *ResourceState, projectPath string, resource *MigrateData) (errors []error) {
	if err := validateMigrateTableRelations(resource.Tables); err != nil {
		errors = append(errors, err)
		return
	}

	wg, errChan, stateChan := sync.WaitGroup{}, make(chan []error), make(chan any)
	doneListen := ListenApplyResource(projectPath, importState, stateChan)

	// role must be run first because will be use when create/update rls
	// and role must be already exist in database
	if len(resource.Roles) > 0 {
		actions := MigrateActionFunc[objects.Role, objects.UpdateRoleParam]{
			CreateFunc: supabase.CreateRole, UpdateFunc: supabase.UpdateRole, DeleteFunc: supabase.DeleteRole,
		}
		errors = runMigrateResource(config, resource.Roles, stateChan, actions, migrateProcess)
		if len(errors) > 0 {
			close(stateChan)
			return errors
		}
	}

	if len(resource.Tables) > 0 {
		cuNode, dNode := BuiltTableTree(resource.Tables)

		actions := MigrateActionFunc[objects.Table, objects.UpdateTableParam]{
			CreateFunc: supabase.CreateTable, UpdateFunc: supabase.UpdateTable,
			DeleteFunc: func(cfg *raiden.Config, param objects.Table) (err error) {
				return supabase.DeleteTable(cfg, param, true)
			},
		}

		logger.Debug(strings.Repeat("=", 5), " Start Migrate Table")
		errors = MigrateTableTree(config, stateChan, actions, cuNode)
		if len(errors) > 0 {
			close(stateChan)
			return errors
		}

		if dNode != nil {
			errors = MigrateTableTree(config, stateChan, actions, dNode)
			if len(errors) > 0 {
				close(stateChan)
				return errors
			}
		}
		logger.Debug(strings.Repeat("=", 5))
	}

	if len(resource.Rpc) > 0 {
		// TODO : handler migrate rpc
		wg.Add(1)
		go func() {
			defer wg.Done()
		}()
	}

	if len(resource.Policies) > 0 {
		// TODO : handler migrate rls
		wg.Add(1)
		go func() {
			defer wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(stateChan)
		close(errChan)
	}()

	for {
		select {
		case rsErr := <-errChan:
			if len(rsErr) > 0 {
				errors = append(errors, rsErr...)
			}
		case saveErr := <-doneListen:
			if saveErr != nil {
				errors = append(errors, saveErr)
			}
			return
		}
	}
}

func runMigrateResource[T any, D any](cfg *raiden.Config, resources []MigrateItem[T, D], stateChan chan any, actionFunc MigrateActionFunc[T, D], migrateFunc MigrateFunc[T, D]) (errors []error) {
	wg, errChan := sync.WaitGroup{}, make(chan error)

	for i := range resources {
		t := resources[i]
		wg.Add(1)

		go func(w *sync.WaitGroup, c *raiden.Config, st chan any, r *MigrateItem[T, D], eChan chan error, acf MigrateActionFunc[T, D]) {
			defer w.Done()
			param := MigrateFuncParam[T, D]{
				Config:      c,
				Data:        t,
				StateChan:   st,
				ActionFuncs: acf,
			}
			err := migrateFunc(param)
			if err != nil {
				eChan <- err
			}
		}(&wg, cfg, stateChan, &t, errChan, actionFunc)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for e := range errChan {
		if e != nil {
			errors = append(errors, e)
		}
	}

	return
}

func migrateProcess[T, D any](params MigrateFuncParam[T, D]) error {
	switch params.Data.Type {
	case MigrateTypeCreate:
		newTable, err := params.ActionFuncs.CreateFunc(params.Config, params.Data.NewData)
		if err != nil {
			return err
		}
		params.Data.NewData = newTable
		params.StateChan <- &params.Data
		return nil
	case MigrateTypeUpdate:
		err := params.ActionFuncs.UpdateFunc(params.Config, params.Data.NewData, params.Data.MigrationItems)
		if err != nil {
			return err
		}

		params.StateChan <- &params.Data
		return nil
	case MigrateTypeDelete:
		err := params.ActionFuncs.DeleteFunc(params.Config, params.Data.OldData)
		if err != nil {
			return err
		}
		params.StateChan <- &params.Data
		return nil
	default:
		return nil
	}
}

func validateMigrateTableRelations(migratedTables []MigrateItem[objects.Table, objects.UpdateTableParam]) error {
	// convert array data to map data
	mapMigratedTableColumns := make(map[string]bool)
	for i := range migratedTables {
		m := migratedTables[i]

		if m.Type == MigrateTypeDelete {
			continue
		}

		if m.NewData.Name != "" {
			for ic := range m.NewData.Columns {
				c := m.NewData.Columns[ic]
				key := fmt.Sprintf("%s.%s", m.NewData.Name, c.Name)
				mapMigratedTableColumns[key] = true
			}
			continue
		}
	}

	for i := range migratedTables {
		m := migratedTables[i]

		if m.Type == MigrateTypeDelete {
			continue
		}

		if m.NewData.Name == "" {
			return fmt.Errorf("validate relation : invalid table name for create or update")
		}

		for i := range m.NewData.Relationships {
			r := m.NewData.Relationships[i]
			if r.SourceTableName != m.NewData.Name {
				continue
			}

			// validate foreign keys
			fkColumn := fmt.Sprintf("%s.%s", r.SourceTableName, r.SourceColumnName)
			if _, exist := mapMigratedTableColumns[fkColumn]; !exist {
				return fmt.Errorf("validate relation table %s : column %s is not exist in table %s", m.NewData.Name, r.SourceColumnName, r.SourceTableName)
			}

			// validate target column
			fkTargetColumn := fmt.Sprintf("%s.%s", r.TargetTableName, r.TargetColumnName)
			if _, exist := mapMigratedTableColumns[fkTargetColumn]; !exist {
				return fmt.Errorf("validate relation table %s : target column %s is not exist in table %s", m.NewData.Name, r.TargetColumnName, r.TargetTableName)
			}
		}
	}

	return nil
}

func MigrateTableTree(cfg *raiden.Config, stateChan chan any, actions MigrateActionFunc[objects.Table, objects.UpdateTableParam], node *MigrateTableNode) []error {
	logger.Debug("start migrate tables level : ", node.Level)
	errors := runMigrateResource(cfg, node.MigratedItems, stateChan, actions, migrateProcess)
	logger.Debug("finish migrate tables level : ", node.Level)
	if len(errors) > 0 {
		return errors
	}

	if node.Child == nil {
		logger.Debug("start migrate child for level : ", node.Level)
		return errors
	}

	return MigrateTableTree(cfg, stateChan, actions, node.Child)
}

// split table base on relation tree and grouping table by level
//
// smallest level is indicate table collection is doesn`t have any relation to other table (master table)
//
// the next level is a table that has a relationship with the table at the previous level
func BuiltTableTree(tables []MigrateItem[objects.Table, objects.UpdateTableParam]) (createAndUpdateNode *MigrateTableNode, deletedNode *MigrateTableNode) {
	logger.Debug(strings.Repeat("=", 5), " BuiltTableTree")
	mapTables := make(map[string]MigrateItem[objects.Table, objects.UpdateTableParam])
	for i := range tables {
		t := tables[i]

		if t.Type == MigrateTypeDelete {
			if deletedNode == nil {
				deletedNode = &MigrateTableNode{
					Level: 1,
				}
			}
			deletedNode.MigratedItems = append(deletedNode.MigratedItems, t)
			continue
		}

		mapTables[t.NewData.Name] = t
	}

	scannedTable := make(map[string]bool)

	createAndUpdateNode = buildTableTree(mapTables, scannedTable, &MigrateTableNode{}, 1)
	logger.Debug(strings.Repeat("=", 5))
	return
}

func buildTableTree(mapTable map[string]MigrateItem[objects.Table, objects.UpdateTableParam], scannedTable map[string]bool, node *MigrateTableNode, level int) (nextNode *MigrateTableNode) {
	scannedNodeTable := make(map[string]bool)

	node.Level = level
	if level == 1 {
		for k, t := range mapTable {
			isMasterTable := true

			// check is master table and add to node members
			if len(t.NewData.Relationships) >= 0 {
				for i := range t.NewData.Relationships {
					r := t.NewData.Relationships[i]
					if r.SourceTableName == t.NewData.Name {
						isMasterTable = false
						break
					}
				}
			}

			logger.Debugf("table %s isMasterTable : %v", t.NewData.Name, isMasterTable)
			if !isMasterTable {
				logger.Debugf("skip processing table %s", t.NewData.Name)
				continue
			}
			node.MigratedItems = append(node.MigratedItems, t)
			key := fmt.Sprintf("%s.%s", t.NewData.Schema, t.NewData.Name)
			logger.Debugf("add %s to scanned table level %v", key, level)
			scannedNodeTable[key] = true
			delete(mapTable, k)
		}
	} else {
		for k, t := range mapTable {
			// check if all depend table is scanned and add to node members
			isAllDependTableScanned := true
			for i := range t.NewData.Relationships {
				r := t.NewData.Relationships[i]
				if r.SourceTableName != t.NewData.Name {
					continue
				}

				key := fmt.Sprintf("%s.%s", r.TargetTableSchema, r.TargetTableName)
				_, isExist := scannedTable[key]
				if !isExist {
					isAllDependTableScanned = false
					break
				}
			}

			logger.Debugf("table %s isAllDependTableScanned : %v", t.NewData.Name, isAllDependTableScanned)
			if !isAllDependTableScanned {
				logger.Debugf("skip processing table %s", t.NewData.Name)
				continue
			}

			node.MigratedItems = append(node.MigratedItems, t)
			key := fmt.Sprintf("%s.%s", t.NewData.Schema, t.NewData.Name)
			logger.Debugf("add %s to scanned table level %v", key, level)
			scannedNodeTable[key] = true
			delete(mapTable, k)
		}
	}

	if len(scannedNodeTable) > 0 {
		for k, v := range scannedNodeTable {
			scannedTable[k] = v
		}
	}

	if len(mapTable) > 0 {
		child := &MigrateTableNode{}
		node.Child = buildTableTree(mapTable, scannedTable, child, level+1)
	}

	return node
}
