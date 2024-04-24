package cloud

import (
	"fmt"
	"strings"
	"sync"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetTables(cfg *raiden.Config, includedSchemas []string, includeColumn bool) ([]objects.Table, error) {
	CloudLogger.Trace("Start - fetching table from supabase")
	q, err := sql.GenerateGetTablesQuery(includedSchemas, includeColumn)
	if err != nil {
		err = fmt.Errorf("failed generate query get table for project id %s : %v", cfg.ProjectId, err)
		return []objects.Table{}, err
	}

	rs, err := ExecuteQuery[[]objects.Table](cfg.SupabaseApiUrl, cfg.ProjectId, q, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		err = fmt.Errorf("get tables error : %s", err)
	}
	CloudLogger.Trace("Finish - fetching table from supabase")
	return rs, err
}

func GetTableByName(cfg *raiden.Config, name, schema string, includeColumn bool) (result objects.Table, err error) {
	CloudLogger.Trace("Start - fetching single table by role from supabase")
	q, err := sql.GenerateGetTableQuery(name, schema, includeColumn)
	if err != nil {
		err = fmt.Errorf("failed generate query get table for project id %s : %v", cfg.ProjectId, err)
		return result, err
	}

	rs, err := ExecuteQuery[[]objects.Table](cfg.SupabaseApiUrl, cfg.ProjectId, q, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		err = fmt.Errorf("get tables error : %s", err)
		return
	}

	if len(rs) == 0 {
		err = fmt.Errorf("get table %s in schema %s is not found", name, schema)
		return
	}
	CloudLogger.Trace("Finish - fetching single table by role from supabase")
	return rs[0], nil
}

func CreateTable(cfg *raiden.Config, newTable objects.Table) (result objects.Table, err error) {
	CloudLogger.Trace("Start - create table", "name", newTable.Name)
	schema := "public"
	if newTable.Schema != "" {
		schema = newTable.Schema
	}

	sql, err := query.BuildCreateTableQuery(newTable)
	if err != nil {
		return result, err
	}

	// execute update
	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return result, fmt.Errorf("create new table %s error : %s", newTable.Name, err)
	}

	CloudLogger.Trace("Finish - create table", "name", newTable.Name)
	return GetTableByName(cfg, newTable.Name, schema, true)
}

func UpdateTable(cfg *raiden.Config, newTable objects.Table, updateItem objects.UpdateTableParam) error {
	CloudLogger.Trace("Start - update table", "name", newTable.Name)
	sql := query.BuildUpdateTableQuery(newTable, updateItem)
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("update tables error : %s", err)
	}

	// execute update column
	if len(updateItem.ChangeColumnItems) > 0 {
		errors := updateColumnFromTable(cfg, updateItem.ChangeColumnItems, newTable.Columns, updateItem.OldData.Columns, newTable.PrimaryKeys)
		if len(errors) > 0 {
			var errMsg []string
			for _, e := range errors {
				errMsg = append(errMsg, e.Error())
			}
			return fmt.Errorf(strings.Join(errMsg, ";"))
		}
	}

	if len(updateItem.ChangeRelationItems) > 0 || updateItem.ForceCreateRelation {
		errors := updateRelations(cfg, updateItem.ChangeRelationItems, newTable.Relationships, updateItem.ForceCreateRelation)
		if len(errors) > 0 {
			var errMsg []string

			for _, e := range errors {
				errMsg = append(errMsg, e.Error())
			}
			return fmt.Errorf(strings.Join(errMsg, ";"))
		}
	}
	CloudLogger.Trace("Finish - update table", "name", newTable.Name)
	return nil
}

func DeleteTable(cfg *raiden.Config, table objects.Table, cascade bool) error {
	CloudLogger.Trace("Start - delete table", "name", table.Name)
	sql := query.BuildDeleteTableQuery(table, true)
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("delete table %s error : %s", table.Name, err)
	}
	CloudLogger.Trace("Finish - delete table", "name", table.Name)
	return nil
}

// ----- update column -----
func updateColumnFromTable(
	cfg *raiden.Config, updateColumns []objects.UpdateColumnItem,
	newColumns []objects.Column, oldColumns []objects.Column, primaryKeys []objects.PrimaryKey,
) []error {
	mapNewColumn := make(map[string]objects.Column)
	mapOldColumn := make(map[string]objects.Column)
	mapIsPrimaryColumn := make(map[string]bool)
	wg := sync.WaitGroup{}
	errors := make([]error, 0)
	errChan := make(chan error)

	for i := range newColumns {
		c := newColumns[i]
		mapNewColumn[c.Name] = c
	}

	for i := range oldColumns {
		c := oldColumns[i]
		mapOldColumn[c.Name] = c
	}

	for i := range primaryKeys {
		pk := primaryKeys[i]
		mapIsPrimaryColumn[pk.Name] = true
	}

	for i := range updateColumns {
		cu := updateColumns[i]
		newColumn := mapNewColumn[cu.Name]
		oldColumn := mapOldColumn[cu.Name]

		wg.Add(1)
		go func(w *sync.WaitGroup, eChan chan error, c *raiden.Config, ui objects.UpdateColumnItem, nc objects.Column, oc objects.Column) {
			defer w.Done()

			var isCreate, isUpdate, isDelete bool
			for _, ut := range cu.UpdateItems {
				switch ut {
				case objects.UpdateColumnNew:
					isCreate = true
				case objects.UpdateColumnDelete:
					isDelete = true
				case objects.UpdateColumnName, objects.UpdateColumnDataType, objects.UpdateColumnUnique, objects.UpdateColumnNullable, objects.UpdateColumnDefaultValue, objects.UpdateColumnIdentity:
					isUpdate = true
				default:
					continue
				}
			}

			if isCreate {
				_, isPrimary := mapIsPrimaryColumn[newColumn.Name]
				if err := CreateColumn(c, newColumn, isPrimary); err != nil {
					eChan <- err
					return
				}
			}

			if isUpdate {
				if err := UpdateColumn(c, oldColumn, nc, ui); err != nil {
					eChan <- err
					return
				}
			}

			if isDelete {
				if err := DeleteColumn(c, oldColumn); err != nil {
					eChan <- err
					return
				}
			}

			eChan <- nil
		}(&wg, errChan, cfg, cu, newColumn, oldColumn)
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
	return errors
}

func CreateColumn(cfg *raiden.Config, column objects.Column, isPrimary bool) error {
	CloudLogger.Trace("Start - create column", "table", column.Table, "name", column.Name)
	if column.Schema == "" {
		column.Schema = "public"
	}

	sql, err := query.BuildCreateColumnQuery(column, isPrimary)
	if err != nil {
		return err
	}

	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("create column %s.%s error : %s", column.Table, column.Name, err)
	}
	CloudLogger.Trace("Finish - create column", "table", column.Table, "name", column.Name)
	return nil
}

func UpdateColumn(cfg *raiden.Config, oldColumn, newColumn objects.Column, updateItem objects.UpdateColumnItem) error {
	CloudLogger.Trace("Start - update column", "table", oldColumn.Table, "name", newColumn.Name)

	sql := query.BuildUpdateColumnQuery(oldColumn, newColumn, updateItem)
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("update column %s.%s error : %s", newColumn.Table, newColumn.Name, err)
	}

	CloudLogger.Trace("Finish - update column", "table", oldColumn.Table, "name", newColumn.Name)
	return nil
}

func DeleteColumn(cfg *raiden.Config, column objects.Column) error {
	CloudLogger.Trace("Start - delete column", "table", column.Table, "name", column.Name)
	sql := query.BuildDeleteColumnQuery(column)
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("delete column %s.%s error : %s", column.Table, column.Name, err)
	}
	CloudLogger.Trace("Finish - delete column", "table", column.Table, "name", column.Name)
	return nil
}

// ----- update relation -----

func updateRelations(cfg *raiden.Config, items []objects.UpdateRelationItem, relations []objects.TablesRelationship, forceCreate bool) []error {
	wg := sync.WaitGroup{}
	errors := make([]error, 0)
	errChan := make(chan error)

	relationMap := make(map[string]*objects.TablesRelationship)
	for i := range relations {
		r := relations[i]
		if r.ConstraintName == "" {
			r.ConstraintName = getRelationConstrainName(r.SourceSchema, r.SourceTableName, r.SourceColumnName)
		}
		relationMap[r.ConstraintName] = &r
	}

	if forceCreate {
		for _, r := range relationMap {
			wg.Add(1)
			go func(w *sync.WaitGroup, c *raiden.Config, r *objects.TablesRelationship, eChan chan error) {
				defer w.Done()
				if err := createForeignKey(cfg, r); err != nil {
					eChan <- err
					return
				}
				eChan <- nil
			}(&wg, cfg, r, errChan)
		}
	} else {
		for _, i := range items {
			switch i.Type {
			case objects.UpdateRelationCreate:
				rel, exist := relationMap[i.Data.ConstraintName]
				if !exist {
					continue
				}

				wg.Add(1)
				go func(w *sync.WaitGroup, c *raiden.Config, r *objects.TablesRelationship, eChan chan error) {
					defer w.Done()
					if err := createForeignKey(cfg, r); err != nil {
						eChan <- err
						return
					}
					eChan <- nil
				}(&wg, cfg, rel, errChan)
			case objects.UpdateRelationUpdate:
				rel, exist := relationMap[i.Data.ConstraintName]
				if !exist {
					continue
				}

				wg.Add(1)
				go func(w *sync.WaitGroup, c *raiden.Config, r *objects.TablesRelationship, eChan chan error) {
					defer w.Done()
					if err := updateForeignKey(cfg, r); err != nil {
						eChan <- err
						return
					}
					eChan <- nil
				}(&wg, cfg, rel, errChan)
			case objects.UpdateRelationDelete:
				wg.Add(1)
				go func(w *sync.WaitGroup, c *raiden.Config, r *objects.TablesRelationship, eChan chan error) {
					defer w.Done()
					if err := deleteForeignKey(cfg, r); err != nil {
						eChan <- err
						return
					}
					eChan <- nil
				}(&wg, cfg, &i.Data, errChan)
			}
		}
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
	return errors
}

func createForeignKey(cfg *raiden.Config, relation *objects.TablesRelationship) error {
	CloudLogger.Trace("Start - create foreign key", "table", relation.TargetTableName, "constrain-name", relation.ConstraintName)

	sql, err := query.BuildFkQuery(objects.UpdateRelationCreate, relation)
	if err != nil {
		return err
	}

	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("create foreign key %s.%s error : %s", relation.SourceTableName, relation.SourceColumnName, err)
	}

	CloudLogger.Trace("Finish - create foreign key", "table", relation.TargetTableName, "constrain-name", relation.ConstraintName)
	return nil
}

func updateForeignKey(cfg *raiden.Config, relation *objects.TablesRelationship) error {
	CloudLogger.Trace("Start - update foreign key", "table", relation.SourceTableName, "constrain-name", relation.ConstraintName)

	deleteSql, err := query.BuildFkQuery(objects.UpdateRelationDelete, relation)
	if err != nil {
		return err
	}

	createSql, err := query.BuildFkQuery(objects.UpdateRelationCreate, relation)
	if err != nil {
		return err
	}

	sql := deleteSql + createSql
	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("update foreign key %s.%s error : %s", relation.SourceTableName, relation.SourceColumnName, err)
	}

	CloudLogger.Trace("FInish - update foreign key", "table", relation.SourceTableName, "constrain-name", relation.ConstraintName)
	return nil
}

func deleteForeignKey(cfg *raiden.Config, relation *objects.TablesRelationship) error {
	CloudLogger.Trace("Start - delete foreign key", "table", relation.SourceTableName, "constrain-name", relation.ConstraintName)

	sql, err := query.BuildFkQuery(objects.UpdateRelationDelete, relation)
	if err != nil {
		return err
	}

	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("delete foreign key %s.%s error : %s", relation.SourceTableName, relation.SourceColumnName, err)
	}

	CloudLogger.Trace("Finish - delete foreign key", "table", relation.SourceTableName, "constrain-name", relation.ConstraintName)
	return nil
}

// get relation table name, base on struct type that defined in relation field
func getRelationConstrainName(schema, table, foreignKey string) string {
	return fmt.Sprintf("%s_%s_%s_fkey", schema, table, foreignKey)
}
