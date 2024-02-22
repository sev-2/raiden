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
	Tables   []MigrateItem[objects.Table, objects.UpdateTableItem]
	Roles    []MigrateItem[objects.Role, any]
	Rpc      []MigrateItem[objects.Function, any]
	Policies []MigrateItem[objects.Policies, any]
}

func MigrateResource(config *raiden.Config, importState *resourceState, projectPath string, resource *MigrateData) (errors []error) {
	wg, errChan, stateChan := sync.WaitGroup{}, make(chan []error), make(chan any)
	doneListen := ListenApplyResource(projectPath, importState, stateChan)

	if len(resource.Tables) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errChan <- migrateTables(config, resource.Tables, stateChan)
		}()
	}

	if len(resource.Roles) > 0 {
		// TODO : handler migrate roles
		wg.Add(1)
		go func() {
			defer wg.Done()
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

func migrateTables(cfg *raiden.Config, tables []MigrateItem[objects.Table, objects.UpdateTableItem], stateChan chan any) (errors []error) {
	wg, errChan := sync.WaitGroup{}, make(chan error)

	for i := range tables {
		t := tables[i]
		wg.Add(1)
		go migrateTable(&wg, cfg, stateChan, &t, errChan)
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

func migrateTable(w *sync.WaitGroup, c *raiden.Config, st chan any, m *MigrateItem[objects.Table, objects.UpdateTableItem], eChan chan error) {
	defer w.Done()

	switch m.Type {
	case MigrateTypeCreate:
		newTable, err := supabase.CreateTable(c, m.NewData)
		if err != nil {
			eChan <- err
			return
		}
		m.NewData = newTable

		st <- m
		eChan <- nil
		return
	case MigrateTypeUpdate:
		err := supabase.UpdateTable(c, m.OldData, m.NewData, m.MigrationItems)
		if err != nil {
			eChan <- err
			return
		}
		st <- m
		eChan <- nil
		return
	case MigrateTypeDelete:
		err := supabase.DeleteTable(c, m.OldData, true)
		if err != nil {
			eChan <- err
			return
		}
		st <- m
		eChan <- nil
		return
	case MigrateTypeIgnore:
		eChan <- nil
		return
	}
}
