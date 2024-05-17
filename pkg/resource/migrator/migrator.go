package migrator

import (
	"sync"

	"github.com/sev-2/raiden"
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

func MigrateResource[T any, D any](cfg *raiden.Config, resources []MigrateItem[T, D], stateChan chan any, actionFunc MigrateActionFunc[T, D], migrateFunc MigrateFunc[T, D]) (errors []error) {
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

func MigratePolicy(cfg *raiden.Config, resources []MigrateItem[objects.Policy, objects.UpdatePolicyParam], stateChan chan any, actionFunc MigrateActionFunc[objects.Policy, objects.UpdatePolicyParam]) (errors []error) {
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

func DefaultMigrator[T, D any](params MigrateFuncParam[T, D]) error {
	switch params.Data.Type {
	case MigrateTypeCreate:
		newData, err := params.ActionFuncs.CreateFunc(params.Config, params.Data.NewData)
		if err != nil {
			return err
		}
		params.Data.NewData = newData
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
