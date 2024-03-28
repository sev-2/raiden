package resource

import (
	"sync"

	"github.com/sev-2/raiden"
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
	Policies []MigrateItem[objects.Policy, objects.UpdatePolicyParam]
	Storages []MigrateItem[objects.Bucket, objects.UpdateBucketParam]
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
		actions := MigrateActionFunc[objects.Table, objects.UpdateTableParam]{
			CreateFunc: supabase.CreateTable, UpdateFunc: supabase.UpdateTable,
			DeleteFunc: func(cfg *raiden.Config, param objects.Table) (err error) {
				return supabase.DeleteTable(cfg, param, true)
			},
		}

		var updateTableRelation []MigrateItem[objects.Table, objects.UpdateTableParam]
		for i := range resource.Tables {
			t := resource.Tables[i]
			if len(t.NewData.Relationships) > 0 {
				if t.Type == MigrateTypeCreate {
					updateTableRelation = append(updateTableRelation, MigrateItem[objects.Table, objects.UpdateTableParam]{
						Type:    MigrateTypeUpdate,
						NewData: t.NewData,
						MigrationItems: objects.UpdateTableParam{
							OldData:             t.NewData,
							ChangeRelationItems: t.MigrationItems.ChangeRelationItems,
							ForceCreateRelation: true,
						},
					})
					resource.Tables[i].MigrationItems.ChangeRelationItems = make([]objects.UpdateRelationItem, 0)
				} else {
					updateTableRelation = append(updateTableRelation, MigrateItem[objects.Table, objects.UpdateTableParam]{
						Type:    t.Type,
						NewData: t.NewData,
						OldData: t.OldData,
						MigrationItems: objects.UpdateTableParam{
							OldData:             t.NewData,
							ChangeRelationItems: t.MigrationItems.ChangeRelationItems,
						},
					})
					resource.Tables[i].MigrationItems.ChangeRelationItems = make([]objects.UpdateRelationItem, 0)
				}
			}
		}

		// run migrate for table manipulation
		errors = runMigrateResource(config, resource.Tables, stateChan, actions, migrateProcess)
		if len(errors) > 0 {
			close(stateChan)
			return errors
		}

		// run migrate for relation manipulation
		if len(updateTableRelation) > 0 {
			errors = runMigrateResource(config, updateTableRelation, stateChan, actions, migrateProcess)
			if len(errors) > 0 {
				close(stateChan)
				return errors
			}
		}
	}

	if len(resource.Rpc) > 0 {
		wg.Add(1)
		go func(w *sync.WaitGroup, eChan chan []error) {
			defer wg.Done()

			actions := MigrateActionFunc[objects.Function, any]{
				CreateFunc: supabase.CreateFunction,
				UpdateFunc: func(cfg *raiden.Config, param objects.Function, items any) (err error) {
					return supabase.UpdateFunction(cfg, param)
				},
				DeleteFunc: supabase.DeleteFunction,
			}

			errors := runMigrateResource(config, resource.Rpc, stateChan, actions, migrateProcess)
			if len(errors) > 0 {
				eChan <- errors
				return
			}
		}(&wg, errChan)
	}

	if len(resource.Policies) > 0 {
		wg.Add(1)
		go func(w *sync.WaitGroup, eChan chan []error) {
			defer w.Done()
			actions := MigrateActionFunc[objects.Policy, objects.UpdatePolicyParam]{
				CreateFunc: supabase.CreatePolicy,
				UpdateFunc: supabase.UpdatePolicy,
				DeleteFunc: supabase.DeletePolicy,
			}

			errors := runMigratePolicy(config, resource.Policies, stateChan, actions)
			if len(errors) > 0 {
				eChan <- errors
				return
			}
		}(&wg, errChan)
	}

	if len(resource.Storages) > 0 {
		wg.Add(1)
		go func(w *sync.WaitGroup, eChan chan []error) {
			defer w.Done()
			actions := MigrateActionFunc[objects.Bucket, objects.UpdateBucketParam]{
				CreateFunc: supabase.CreateBucket,
				UpdateFunc: supabase.UpdateBucket,
				DeleteFunc: supabase.DeleteBucket,
			}

			errors := runMigrateResource(config, resource.Storages, stateChan, actions, migrateProcess)
			if len(errors) > 0 {
				eChan <- errors
				return
			}
		}(&wg, errChan)
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

func runMigratePolicy(cfg *raiden.Config, resources []MigrateItem[objects.Policy, objects.UpdatePolicyParam], stateChan chan any, actionFunc MigrateActionFunc[objects.Policy, objects.UpdatePolicyParam]) (errors []error) {
	mapPoliciesByTable := make(map[string][]MigrateItem[objects.Policy, objects.UpdatePolicyParam])
	for i := range resources {
		p := resources[i]

		var tableName string
		if p.NewData.Name != "" {
			tableName = p.NewData.Table
		} else if p.OldData.Name != "" {
			tableName = p.OldData.Table
		}

		policies, exist := mapPoliciesByTable[tableName]
		if !exist {
			mapPoliciesByTable[tableName] = []MigrateItem[objects.Policy, objects.UpdatePolicyParam]{
				p,
			}
			continue
		}
		policies = append(policies, p)
		mapPoliciesByTable[tableName] = policies
	}

	wg, errChan := sync.WaitGroup{}, make(chan error)

	for _, v := range mapPoliciesByTable {
		wg.Add(1)
		go func(
			w *sync.WaitGroup, c *raiden.Config, st chan any,
			mList []MigrateItem[objects.Policy, objects.UpdatePolicyParam],
			acf MigrateActionFunc[objects.Policy, objects.UpdatePolicyParam],
			eChan chan error,
		) {
			defer w.Done()

			for i := range mList {
				m := mList[i]

				switch m.Type {
				case MigrateTypeCreate:
					newTable, err := acf.CreateFunc(c, m.NewData)
					if err != nil {
						eChan <- err
						continue
					}
					m.NewData = newTable
					stateChan <- &m
				case MigrateTypeUpdate:
					err := acf.UpdateFunc(c, m.NewData, m.MigrationItems)
					if err != nil {
						eChan <- err
						continue
					}
					stateChan <- &m
				case MigrateTypeDelete:
					err := acf.DeleteFunc(c, m.OldData)
					if err != nil {
						eChan <- err
						continue
					}
					stateChan <- &m
				}
			}
		}(&wg, cfg, stateChan, v, actionFunc, errChan)
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
