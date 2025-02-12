package pgmeta

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/client/net"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetTables(cfg *raiden.Config, includedSchemas []string, includeColumns bool) ([]objects.Table, error) {
	MetaLogger.Trace("start fetching tables from meta")
	url := fmt.Sprintf("%s/tables", cfg.PgMetaUrl)

	reqInterceptor := func(req *http.Request) error {
		if len(includedSchemas) > 0 {
			req.URL.Query().Set("included_schemas", strings.Join(includedSchemas, ","))
		}

		if includeColumns {
			req.URL.Query().Set("include_columns", strconv.FormatBool(includeColumns))
		}

		if len(cfg.JwtToken) > 0 {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.JwtToken))
		}

		return nil
	}

	rs, err := net.Get[[]objects.Table](url, net.DefaultTimeout, reqInterceptor, nil)
	if err != nil {
		err = fmt.Errorf("get tables error : %s", err)
	}
	MetaLogger.Trace("finish fetching tables from meta")
	return rs, err
}

func GetTableByName(cfg *raiden.Config, name, schema string, includeColumn bool) (result objects.Table, err error) {
	MetaLogger.Trace("start fetching table by name from meta")
	q, err := sql.GenerateGetTableQuery(name, schema, includeColumn)
	if err != nil {
		err = fmt.Errorf("failed generate query get table for project id %s : %v", cfg.ProjectId, err)
		return result, err
	}

	rs, err := ExecuteQuery[[]objects.Table](cfg.PgMetaUrl, q, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		err = fmt.Errorf("get tables error : %s", err)
		return
	}

	if len(rs) == 0 {
		err = fmt.Errorf("get table %s in schema %s is not found", name, schema)
		return
	}
	MetaLogger.Trace("finish fetching table by name from meta")
	return rs[0], nil
}

func CreateTable(cfg *raiden.Config, newTable objects.Table) (result objects.Table, err error) {
	MetaLogger.Trace("start create table", "name", newTable.Name)
	schema := "public"
	if newTable.Schema != "" {
		schema = newTable.Schema
	}

	sql, err := query.BuildCreateTableQuery(newTable)
	if err != nil {
		return result, err
	}

	// execute update
	_, err = ExecuteQuery[any](cfg.PgMetaUrl, sql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		return result, fmt.Errorf("create new table %s error : %s", newTable.Name, err)
	}
	MetaLogger.Trace("finish create table", "name", newTable.Name)
	return GetTableByName(cfg, newTable.Name, schema, true)
}

func UpdateTable(cfg *raiden.Config, newTable objects.Table, updateItem objects.UpdateTableParam) error {
	MetaLogger.Trace("start update table", "name", newTable.Name)
	sql := query.BuildUpdateTableQuery(newTable, updateItem)
	// execute update
	_, err := ExecuteQuery[any](cfg.PgMetaUrl, sql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		return fmt.Errorf("update tables error : %s", err)
	}

	// execute update column
	if len(updateItem.ChangeColumnItems) > 0 {
		errorsMessage := updateColumnFromTable(cfg, updateItem.ChangeColumnItems, newTable.Columns, updateItem.OldData.Columns, newTable.PrimaryKeys)
		if len(errorsMessage) > 0 {
			var errMsg []string
			for _, e := range errorsMessage {
				errMsg = append(errMsg, e.Error())
			}
			return errors.New(strings.Join(errMsg, ";"))
		}
	}
	if len(updateItem.ChangeRelationItems) > 0 || updateItem.ForceCreateRelation {
		errorsUpdate := updateRelations(cfg, updateItem.ChangeRelationItems, newTable.Relationships, updateItem.ForceCreateRelation)
		if len(errorsUpdate) > 0 {
			var errMsg []string

			for _, e := range errorsUpdate {
				errMsg = append(errMsg, e.Error())
			}
			return errors.New(strings.Join(errMsg, ";"))
		}
	}
	MetaLogger.Trace("finish update table", "name", newTable.Name)
	return nil
}

func DeleteTable(cfg *raiden.Config, table objects.Table, cascade bool) error {
	MetaLogger.Trace("start delete table", "name", table.Name)
	sql := query.BuildDeleteTableQuery(table, true)
	// execute delete
	_, err := ExecuteQuery[any](cfg.PgMetaUrl, sql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		return fmt.Errorf("delete table %s error : %s", table.Name, err)
	}
	MetaLogger.Trace("finish delete table", "name", table.Name)
	return nil
}

// ----- relationship action -----
func GetTableRelationshipActions(cfg *raiden.Config, schema string) ([]objects.TablesRelationshipAction, error) {
	MetaLogger.Trace("start fetching table relationships from pg-meta")
	q := sql.GenerateGetTableRelationshipActionsQuery(schema)

	rs, err := ExecuteQuery[[]objects.TablesRelationshipAction](cfg.PgMetaUrl, q, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		return rs, fmt.Errorf("get tables relation from pg-meta error : %s", err)
	}

	MetaLogger.Trace("finish fetching table relationships from pg-meta")
	return rs, nil
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
	MetaLogger.Trace("start create column", "table", column.Table, "name", column.Name)
	if column.Schema == "" {
		column.Schema = "public"
	}

	sql, err := query.BuildCreateColumnQuery(column, isPrimary)
	if err != nil {
		return err
	}

	// Execute SQL Query
	_, err = ExecuteQuery[any](cfg.PgMetaUrl, sql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		return fmt.Errorf("create column %s.%s error : %s", column.Table, column.Name, err)
	}
	MetaLogger.Trace("finish create column", "table", column.Table, "name", column.Name)
	return nil
}

func UpdateColumn(cfg *raiden.Config, oldColumn, newColumn objects.Column, updateItem objects.UpdateColumnItem) error {
	MetaLogger.Trace("start update column", "table", oldColumn.Table, "name", oldColumn.Name)
	// Build Execute Query
	sql := query.BuildUpdateColumnQuery(oldColumn, newColumn, updateItem)

	// Execute SQL Query
	_, err := ExecuteQuery[any](cfg.PgMetaUrl, sql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		return fmt.Errorf("update column %s.%s error : %s", newColumn.Table, newColumn.Name, err)
	}
	MetaLogger.Trace("finish update column", "table", oldColumn.Table, "name", oldColumn.Name)
	return nil
}

func DeleteColumn(cfg *raiden.Config, column objects.Column) error {
	MetaLogger.Trace("start delete column", "table", column.Table, "name", column.Name)
	sql := query.BuildDeleteColumnQuery(column)
	_, err := ExecuteQuery[any](cfg.PgMetaUrl, sql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		return fmt.Errorf("delete column %s.%s error : %s", column.Table, column.Name, err)
	}
	MetaLogger.Trace("finish delete column", "table", column.Table, "name", column.Name)
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
			case objects.UpdateRelationUpdate, objects.UpdateRelationActionOnDelete, objects.UpdateRelationActionOnUpdate, objects.UpdateRelationCreateIndex:
				rel, exist := relationMap[i.Data.ConstraintName]
				if !exist {
					actionKey := fmt.Sprintf("%s_%s_fkey", i.Data.SourceTableName, i.Data.SourceColumnName)
					rel2, exist2 := relationMap[actionKey]
					if !exist2 {
						continue
					}
					rel = rel2
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
	MetaLogger.Trace("start create foreign key", "table", relation.TargetTableName, "constrain-name", relation.ConstraintName)
	sql, err := query.BuildFkQuery(objects.UpdateRelationCreate, relation)
	if err != nil {
		return err
	}

	_, err = ExecuteQuery[any](cfg.PgMetaUrl, sql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		return fmt.Errorf("create foreign key %s.%s error : %s", relation.SourceTableName, relation.SourceColumnName, err)
	}

	// create FK index
	if indexSql, err := query.BuildFKIndexQuery(objects.UpdateRelationCreate, relation); err != nil {
		return err
	} else if len(indexSql) > 0 {
		_, err = ExecuteQuery[any](cfg.PgMetaUrl, indexSql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
		if err != nil {
			return fmt.Errorf("create foreign index %s.%s error : %s", relation.SourceTableName, relation.SourceColumnName, err)
		}
	}

	MetaLogger.Trace("finish create foreign key", "table", relation.TargetTableName, "constrain-name", relation.ConstraintName)
	return nil
}

func updateForeignKey(cfg *raiden.Config, relation *objects.TablesRelationship) error {
	MetaLogger.Trace("start update foreign key", "table", relation.TargetTableName, "constrain-name", relation.ConstraintName)
	deleteSql, err := query.BuildFkQuery(objects.UpdateRelationDelete, relation)
	if err != nil {
		return err
	}

	createSql, err := query.BuildFkQuery(objects.UpdateRelationCreate, relation)
	if err != nil {
		return err
	}

	sql := deleteSql + createSql
	_, err = ExecuteQuery[any](cfg.PgMetaUrl, sql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		return fmt.Errorf("update foreign key %s.%s error : %s", relation.SourceTableName, relation.SourceColumnName, err)
	}

	if relation.Index == nil {
		// create FK index
		if indexSql, err := query.BuildFKIndexQuery(objects.UpdateRelationCreate, relation); err != nil {
			return err
		} else if len(indexSql) > 0 {
			_, err = ExecuteQuery[any](cfg.PgMetaUrl, indexSql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
			if err != nil {
				return fmt.Errorf("create foreign index %s.%s error : %s", relation.SourceTableName, relation.SourceColumnName, err)
			}
		}
	}

	MetaLogger.Trace("finish update foreign key", "table", relation.TargetTableName, "constrain-name", relation.ConstraintName)
	return nil
}

func deleteForeignKey(cfg *raiden.Config, relation *objects.TablesRelationship) error {
	MetaLogger.Trace("start delete foreign key", "table", relation.TargetTableName, "constrain-name", relation.ConstraintName)

	sql, err := query.BuildFkQuery(objects.UpdateRelationDelete, relation)
	if err != nil {
		return err
	}

	_, err = ExecuteQuery[any](cfg.PgMetaUrl, sql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		return fmt.Errorf("delete foreign key %s.%s error : %s", relation.SourceTableName, relation.SourceColumnName, err)
	}

	if relation.Index != nil {
		// create FK index
		if indexSql, err := query.BuildFKIndexQuery(objects.UpdateRelationDelete, relation); err != nil {
			return err
		} else if len(indexSql) > 0 {
			_, err = ExecuteQuery[any](cfg.PgMetaUrl, indexSql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
			if err != nil {
				return fmt.Errorf("delete foreign index %s.%s error : %s", relation.SourceTableName, relation.SourceColumnName, err)
			}
		}
	}

	MetaLogger.Trace("start delete foreign key", "table", relation.TargetTableName, "constrain-name", relation.ConstraintName)
	return nil
}

// get relation table name, base on struct type that defined in relation field
func getRelationConstrainName(schema, table, foreignKey string) string {
	return fmt.Sprintf("%s_%s_%s_fkey", schema, table, foreignKey)
}
