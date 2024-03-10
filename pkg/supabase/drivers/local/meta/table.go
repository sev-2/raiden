package meta

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/client"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
	"github.com/valyala/fasthttp"
)

func GetTables(cfg *raiden.Config, includedSchemas []string, includeColumns bool) ([]objects.Table, error) {
	url := fmt.Sprintf("%s%s/tables", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
	reqInterceptor := func(req *fasthttp.Request) error {
		if len(includedSchemas) > 0 {
			req.URI().QueryArgs().Set("included_schemas", strings.Join(includedSchemas, ","))
		}

		if includeColumns {
			req.URI().QueryArgs().Set("include_columns", strconv.FormatBool(includeColumns))
		}

		return nil
	}

	rs, err := client.Get[[]objects.Table](url, client.DefaultTimeout, reqInterceptor, nil)
	if err != nil {
		err = fmt.Errorf("get tables error : %s", err)
	}
	return rs, err
}

func GetTableByName(cfg *raiden.Config, name, schema string, includeColumn bool) (result objects.Table, err error) {
	q, err := sql.GenerateGetTableQuery(name, schema, includeColumn)
	if err != nil {
		err = fmt.Errorf("failed generate query get table for project id %s : %v", cfg.ProjectId, err)
		return result, err
	}

	rs, err := ExecuteQuery[[]objects.Table](getBaseUrl(cfg), q, nil, nil, nil)
	if err != nil {
		err = fmt.Errorf("get tables error : %s", err)
		return
	}

	if len(rs) == 0 {
		err = fmt.Errorf("get table %s in schema %s is not found", name, schema)
		return
	}

	return rs[0], nil
}

func CreateTable(cfg *raiden.Config, newTable objects.Table) (result objects.Table, err error) {
	schema := "public"
	if newTable.Schema != "" {
		schema = newTable.Schema
	}

	sql, err := query.BuildCreateTableQuery(newTable)
	if err != nil {
		return result, err
	}

	// execute update
	logger.Debug("Create Table - execute : ", sql)
	_, err = ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return result, fmt.Errorf("create new table %s error : %s", newTable.Name, err)
	}
	return GetTableByName(cfg, newTable.Name, schema, true)
}

func UpdateTable(cfg *raiden.Config, newTable objects.Table, updateItem objects.UpdateTableParam) error {
	sql := query.BuildUpdateTableQuery(newTable, updateItem)
	// execute update
	logger.Debug("Update Table - execute : ", sql)
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
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

	if len(updateItem.ChangeRelationItems) > 0 {
		errors := updateRelations(cfg, updateItem.ChangeRelationItems, newTable.Relationships)
		if len(errors) > 0 {
			var errMsg []string

			for _, e := range errors {
				errMsg = append(errMsg, e.Error())
			}
			return fmt.Errorf(strings.Join(errMsg, ";"))
		}
	}

	return nil
}

func DeleteTable(cfg *raiden.Config, table objects.Table, cascade bool) error {
	sql := query.BuildDeleteTableQuery(table, true)
	// execute delete
	logger.Debug("Delete Table - execute : ", sql)
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("delete table %s error : %s", table.Name, err)
	}

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

func UpdateColumn(cfg *raiden.Config, oldColumn, newColumn objects.Column, updateItem objects.UpdateColumnItem) error {
	// Build Execute Query
	sql := query.BuildUpdateColumnQuery(oldColumn, newColumn, updateItem)

	// Execute SQL Query
	logger.Debug("Update Column - execute : ", sql)
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("update column %s.%s error : %s", newColumn.Table, newColumn.Name, err)
	}

	return nil
}

func CreateColumn(cfg *raiden.Config, column objects.Column, isPrimary bool) error {
	if column.Schema == "" {
		column.Schema = "public"
	}

	sql, err := query.BuildCreateColumnQuery(column, isPrimary)
	if err != nil {
		return err
	}

	// Execute SQL Query
	logger.Debug("Create Column - execute : ", sql)
	_, err = ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("create column %s.%s error : %s", column.Table, column.Name, err)
	}

	return nil
}

func DeleteColumn(cfg *raiden.Config, column objects.Column) error {
	sql := query.BuildDeleteColumnQuery(column)
	logger.Debug("Delete Column - execute : ", sql)
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("delete column %s.%s error : %s", column.Table, column.Name, err)
	}

	return nil
}

// ----- update relation -----

func updateRelations(cfg *raiden.Config, items []objects.UpdateRelationItem, relations []objects.TablesRelationship) []error {
	relationMap := make(map[string]*objects.TablesRelationship)
	for i := range relations {
		r := relations[i]
		relationMap[r.ConstraintName] = &r
	}

	wg := sync.WaitGroup{}
	errors := make([]error, 0)
	errChan := make(chan error)
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
	sql, err := query.BuildFkQuery(objects.UpdateRelationCreate, relation)
	if err != nil {
		return err
	}

	logger.Debug("Create foreign key - execute : ", sql)
	_, err = ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("create foreign key %s.%s error : %s", relation.SourceTableName, relation.SourceColumnName, err)
	}

	return nil
}

func updateForeignKey(cfg *raiden.Config, relation *objects.TablesRelationship) error {
	deleteSql, err := query.BuildFkQuery(objects.UpdateRelationDelete, relation)
	if err != nil {
		return err
	}

	createSql, err := query.BuildFkQuery(objects.UpdateRelationCreate, relation)
	if err != nil {
		return err
	}

	sql := deleteSql + createSql
	logger.Debug("Update foreign key - execute : ", sql)
	_, err = ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("update foreign key %s.%s error : %s", relation.SourceTableName, relation.SourceColumnName, err)
	}

	return nil
}

func deleteForeignKey(cfg *raiden.Config, relation *objects.TablesRelationship) error {
	sql, err := query.BuildFkQuery(objects.UpdateRelationDelete, relation)
	if err != nil {
		return err
	}
	logger.Debug("Delete foreign key - execute : ", sql)
	_, err = ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("delete foreign key %s.%s error : %s", relation.SourceTableName, relation.SourceColumnName, err)
	}
	return nil
}
