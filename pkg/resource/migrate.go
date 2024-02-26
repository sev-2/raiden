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
	ErrChan     chan error
	Data        MigrateItem[T, D]
	ActionFuncs MigrateActionFunc[T, D]
}
type MigrateFunc[T, D any] func(param MigrateFuncParam[T, D])

func MigrateResource(config *raiden.Config, importState *ResourceState, projectPath string, resource *MigrateData) (errors []error) {
	wg, errChan, stateChan := sync.WaitGroup{}, make(chan []error), make(chan any)
	doneListen := ListenApplyResource(projectPath, importState, stateChan)

	// role must be run first because will be use when create/update rls
	// and role must be already exist in database
	if len(resource.Roles) > 0 {
		actions := MigrateActionFunc[objects.Role, objects.UpdateRoleParam]{
			CreateFunc: supabase.CreateRole, UpdateFunc: supabase.UpdateRole, DeleteFunc: supabase.DeleteRole,
		}
		errors = runMigrateResource[objects.Role, objects.UpdateRoleParam](config, resource.Roles, stateChan, actions, migrateProcess)
		if len(errors) > 0 {
			close(stateChan)
			return errors
		}
	}

	if len(resource.Tables) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			actions := MigrateActionFunc[objects.Table, objects.UpdateTableParam]{
				CreateFunc: supabase.CreateTable, UpdateFunc: supabase.UpdateTable,
				DeleteFunc: func(cfg *raiden.Config, param objects.Table) (err error) {
					return supabase.DeleteTable(cfg, param, true)
				},
			}
			errChan <- runMigrateResource[objects.Table, objects.UpdateTableParam](config, resource.Tables, stateChan, actions, migrateProcess)
		}()
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
				ErrChan:     eChan,
				ActionFuncs: acf,
			}
			migrateFunc(param)
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

func migrateProcess[T, D any](params MigrateFuncParam[T, D]) {
	switch params.Data.Type {
	case MigrateTypeCreate:
		newTable, err := params.ActionFuncs.CreateFunc(params.Config, params.Data.NewData)
		if err != nil {
			params.ErrChan <- err
			return
		}
		params.Data.NewData = newTable

		params.StateChan <- &params.Data
		params.ErrChan <- nil
		return
	case MigrateTypeUpdate:
		err := params.ActionFuncs.UpdateFunc(params.Config, params.Data.NewData, params.Data.MigrationItems)
		if err != nil {
			params.ErrChan <- err
			return
		}

		params.StateChan <- &params.Data
		params.ErrChan <- nil
		return
	case MigrateTypeDelete:
		err := params.ActionFuncs.DeleteFunc(params.Config, params.Data.OldData)
		if err != nil {
			params.ErrChan <- err
			return
		}
		params.StateChan <- &params.Data
		params.ErrChan <- nil
		return
	case MigrateTypeIgnore:
		params.ErrChan <- nil
		return
	}
}
