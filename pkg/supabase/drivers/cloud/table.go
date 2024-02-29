package cloud

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/drivers/cloud/query"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func GetTables(cfg *raiden.Config, includedSchemas []string, includeColumn bool) ([]objects.Table, error) {
	q, err := query.GenerateTablesQuery(includedSchemas, includeColumn)
	if err != nil {
		err = fmt.Errorf("failed generate query get table for project id %s : %v", cfg.ProjectId, err)
		return []objects.Table{}, err
	}

	rs, err := ExecuteQuery[[]objects.Table](cfg.SupabaseApiUrl, cfg.ProjectId, q, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		err = fmt.Errorf("get tables error : %s", err)
	}

	return rs, err
}

func GetTableBy(cfg *raiden.Config, name, schema string, includeColumn bool) (result objects.Table, err error) {
	q, err := query.GenerateTableQuery(name, schema, includeColumn)
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

	return rs[0], nil
}

func CreateTable(cfg *raiden.Config, newTable objects.Table) (result objects.Table, err error) {
	schema := "public"
	if newTable.Schema != "" {
		schema = newTable.Schema
	}

	createSql, err := BuildCreateQuery(schema, newTable)
	if err != nil {
		return result, err
	}

	var rlsEnableQuery string
	if newTable.RLSEnabled {
		rlsEnableQuery = fmt.Sprintf("ALTER TABLE %s.%s ENABLE ROW LEVEL SECURITY;", newTable.Schema, newTable.Name)
	}

	var rlsForcedQuery string
	if newTable.RLSForced {
		rlsForcedQuery = fmt.Sprintf("ALTER TABLE %s.%s FORCE ROW LEVEL SECURITY;", newTable.Schema, newTable.Name)
	}

	sql := fmt.Sprintf(`
	BEGIN;
	  %s
	  %s
	  %s
	COMMIT;
	`, createSql, rlsEnableQuery, rlsForcedQuery)

	// execute update
	logger.Debug("Create Table - execute : ", sql)
	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return result, fmt.Errorf("create new table %s error : %s", newTable.Name, err)
	}

	return GetTableBy(cfg, newTable.Name, schema, true)
}

func UpdateTable(cfg *raiden.Config, newTable objects.Table, updateItem objects.UpdateTableParam) error {
	var enableRlsQuery, forceRlsQuery, primaryKeysQuery, replicaIdentityQuery, schemaQuery, nameQuery string
	alter := fmt.Sprintf("ALTER TABLE %s.%s", updateItem.OldData.Schema, updateItem.OldData.Name)
	for _, uType := range updateItem.ChangeItems {
		switch uType {
		case objects.UpdateTableSchema:
			schemaQuery = fmt.Sprintf("%s SET SCHEMA %s;", alter, newTable.Schema)
		case objects.UpdateTableName:
			if newTable.Name != "" {
				nameQuery = fmt.Sprintf("%s RENAME TO %s;", alter, newTable.Name)
			}
		case objects.UpdateTableRlsEnable:
			if newTable.RLSEnabled {
				enableRlsQuery = fmt.Sprintf("%s ENABLE ROW LEVEL SECURITY;", alter)
			} else {
				enableRlsQuery = fmt.Sprintf("%s DISABLE ROW LEVEL SECURITY;", alter)
			}
		case objects.UpdateTableRlsForced:
			if newTable.RLSForced {
				enableRlsQuery = fmt.Sprintf("%s FORCE ROW LEVEL SECURITY;", alter)
			} else {
				enableRlsQuery = fmt.Sprintf("%s NO FORCE ROW LEVEL SECURITY;", alter)
			}
		case objects.UpdateTableReplicaIdentity:
			// TODO : implement if needed
		case objects.UpdateTablePrimaryKey:
			if len(updateItem.OldData.PrimaryKeys) > 0 {
				primaryKeysQuery += fmt.Sprintf(`
				DO $$
				DECLARE
				  r record;
				BEGIN
				  SELECT conname
					INTO r
					FROM pg_constraint
					WHERE contype = 'p' AND conrelid = %d;
				  EXECUTE %s || quote_ident(r.conname);
				END
				$$;
				`, updateItem.OldData.ID, fmt.Sprintf("%s DROP CONSTRAINT ", alter))
			}

			if len(newTable.PrimaryKeys) > 0 {
				var pkArr []string
				for _, v := range newTable.PrimaryKeys {
					pkArr = append(pkArr, v.Name)
					primaryKeysQuery += fmt.Sprintf("%s ADD PRIMARY KEY (%s);", alter, strings.Join(pkArr, ","))
				}

			}
		}
	}

	sql := fmt.Sprintf(`
	BEGIN;
	  %s
	  %s
	  %s
	  %s
	  %s
	  %s
	COMMIT;
	`, enableRlsQuery, forceRlsQuery, replicaIdentityQuery, primaryKeysQuery, schemaQuery, nameQuery)

	// execute update
	logger.Debug("Update Table - execute : ", sql)
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
	sql := fmt.Sprintf("DROP TABLE %s.%s", table.Schema, table.Name)
	if cascade {
		sql += " CASCADE"
	} else {
		sql += " RESTRICT"
	}
	sql += ";"

	// execute delete
	logger.Debug("Delete Table - execute : ", sql)
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
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
	// Prepare SQL statements
	var sqlStatements []string
	var alter = fmt.Sprintf("ALTER TABLE %s.%s", newColumn.Schema, newColumn.Table)
	for _, uType := range updateItem.UpdateItems {
		switch uType {
		case objects.UpdateColumnName:
			if newColumn.Name != "" {
				sqlStatements = append(
					sqlStatements,
					fmt.Sprintf(
						"%s RENAME COLUMN %s TO %s;", alter, newColumn.Name, newColumn.Name,
					),
				)
			}
		case objects.UpdateColumnDataType:
			sqlStatements = append(
				sqlStatements,
				fmt.Sprintf(
					"%s ALTER COLUMN %s SET DATA TYPE %s USING %s::%s;", alter, oldColumn.Name, newColumn.DataType, oldColumn.Name, newColumn.DataType,
				),
			)
		case objects.UpdateColumnUnique:
			if newColumn.IsUnique {
				sqlStatements = append(
					sqlStatements,
					fmt.Sprintf(
						"%s ADD CONSTRAINT %s UNIQUE (%s);", alter, fmt.Sprintf("%s_%s_unique", newColumn.Table, newColumn.Name), newColumn.Name),
				)
			} else {
				sqlStatements = append(
					sqlStatements,
					fmt.Sprintf(
						"%s DROP CONSTRAINT %s;", alter, fmt.Sprintf("%s_%s_unique", newColumn.Table, newColumn.Name),
					),
				)
			}
		case objects.UpdateColumnNullable:
			if newColumn.IsNullable {
				sqlStatements = append(
					sqlStatements,
					fmt.Sprintf(
						"%s ALTER COLUMN %s DROP NOT NULL;", alter, newColumn.Name,
					),
				)
			} else {
				sqlStatements = append(
					sqlStatements,
					fmt.Sprintf(
						"%s ALTER COLUMN %s SET NOT NULL;", alter, newColumn.Name,
					),
				)
			}
		case objects.UpdateColumnDefaultValue:
			rv := reflect.ValueOf(newColumn.DefaultValue)
			if (rv.Kind() == reflect.Ptr && rv.IsNil()) || rv.Kind() == reflect.Invalid {
				sqlStatements = append(
					sqlStatements,
					fmt.Sprintf(
						"%s ALTER COLUMN %s DROP DEFAULT;", alter, newColumn.Name,
					),
				)
				continue
			}

			var value string
			switch v := newColumn.DefaultValue.(type) {
			case string:
				value = v
			case *string:
				if v != nil {
					value = *v
				}
			}

			defaultValue := fmt.Sprintf("'%v'", value)
			if _, e := strconv.ParseInt(value, 10, 64); e == nil {
				defaultValue = value
			} else if _, e := strconv.ParseUint(value, 10, 64); e == nil {
				defaultValue = value
			} else if _, e := strconv.ParseBool(value); e == nil {
				defaultValue = value
			} else if _, e := strconv.ParseFloat(value, 64); e == nil {
				defaultValue = value
			} else if strings.Contains(value, "()") {
				defaultValue = value
			}

			sqlStatements = append(
				sqlStatements,
				fmt.Sprintf(
					"%s ALTER COLUMN %s SET DEFAULT %s;", alter, newColumn.Name, defaultValue,
				),
			)

		case objects.UpdateColumnIdentity:
			if newColumn.IsIdentity {
				sqlStatements = append(
					sqlStatements,
					fmt.Sprintf(
						"%s ALTER COLUMN %s ADD GENERATED %s AS IDENTITY;", alter, newColumn.Name, newColumn.IdentityGeneration,
					),
				)
			} else {
				sqlStatements = append(
					sqlStatements,
					fmt.Sprintf(
						"%s ALTER COLUMN %s DROP IDENTITY IF EXISTS;", alter, newColumn.Name,
					),
				)
			}
		}
	}

	// Build Execute Query
	sql := "BEGIN;"
	for _, stmt := range sqlStatements {
		sql += " " + stmt
	}
	sql += " COMMIT;"

	// Execute SQL Query
	logger.Debug("Update Column - execute : ", sql)
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("update column %s.%s error : %s", newColumn.Table, newColumn.Name, err)
	}

	return nil
}

func CreateColumn(cfg *raiden.Config, column objects.Column, isPrimary bool) error {
	schema := column.Schema
	if schema == "" {
		schema = "public"
	}

	colDef, err := BuildColumnDef(column)
	if err != nil {
		return err
	}

	isPrimaryKeyClause := ""
	if isPrimary {
		isPrimaryKeyClause = "PRIMARY KEY"
	}

	// TODO : implement check setup
	// checkSql := ""
	// if column.Check != nil {
	// 	checkSql = fmt.Sprintf("CHECK (%s)", *column.Check)
	// }

	// TODO : implement comment setup
	// commentSql := ""
	// if column.Comment != nil {
	// 	commentSql = fmt.Sprintf("COMMENT ON COLUMN %s.%s.%s IS %s", ident(schema), ident(table.Name), ident(name), literal(*column.Comment))
	// }

	sql := fmt.Sprintf(`
	BEGIN;
	  ALTER TABLE %s.%s ADD COLUMN %s %s;
	COMMIT;`, schema, column.Table, colDef, isPrimaryKeyClause)

	// Execute SQL Query
	logger.Debug("Create Column - execute : ", sql)
	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("create column %s.%s error : %s", column.Table, column.Name, err)
	}

	return nil
}

func DeleteColumn(cfg *raiden.Config, column objects.Column) error {
	sql := fmt.Sprintf("ALTER TABLE %s.%s DROP COLUMN %s;", column.Schema, column.Table, column.Name)
	logger.Debug("Delete Column - execute : ", sql)
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("delete column %s.%s error : %s", column.Table, column.Name, err)
	}

	return nil
}

// ----- update relation -----
func BuildCreateQuery(schema string, table objects.Table) (q string, err error) {
	var tableContains []string

	// add column definition
	for i := range table.Columns {
		c := table.Columns[i]
		colDef, err := BuildColumnDef(c)
		if err != nil {
			return q, fmt.Errorf("err build column definition %s : %s", c.Name, err.Error())
		}
		tableContains = append(tableContains, colDef)
	}

	// append primary key
	var primaryKeys []string
	for _, pk := range table.PrimaryKeys {
		primaryKeys = append(primaryKeys, pk.Name)
	}

	if len(primaryKeys) > 0 {
		tableContains = append(tableContains, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(primaryKeys, ",")))
	}

	// append relation
	for _, rel := range table.Relationships {
		if table.Name != rel.SourceTableName {
			continue
		}

		if rel.ConstraintName == "" {
			rel.ConstraintName = fmt.Sprintf("%s_%s_%s_fkey", rel.SourceSchema, rel.SourceTableName, rel.SourceColumnName)
		}

		fkString := fmt.Sprintf("CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s.%s(%s)",
			rel.ConstraintName,
			rel.SourceColumnName,
			rel.TargetTableSchema,
			rel.TargetTableName,
			rel.TargetColumnName,
		)
		tableContains = append(tableContains, fkString)
	}

	q = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (%s);", schema, table.Name, strings.Join(tableContains, ","))
	return
}

func BuildColumnDef(column objects.Column) (string, error) {
	var defaultValueClause string
	if column.IsIdentity {
		if column.DefaultValue != nil {
			return "", fmt.Errorf("columns %s.%s %s cannot both be identity and have a default value", column.Schema, column.Table, column.Name)
		}
		defaultValueClause = fmt.Sprintf("GENERATED %s AS IDENTITY", column.IdentityGeneration)
	} else {
		rv := reflect.ValueOf(column.DefaultValue)
		if (rv.Kind() == reflect.Ptr && rv.IsNil()) || rv.Kind() == reflect.Invalid {
			defaultValueClause = ""
		}

		var value string
		switch v := column.DefaultValue.(type) {
		case string:
			value = v
		case *string:
			value = *v
		}

		if value != "" {
			defaultValue := fmt.Sprintf("'%v'", value)
			if _, e := strconv.ParseInt(value, 10, 64); e == nil {
				defaultValue = value
			} else if _, e := strconv.ParseUint(value, 10, 64); e == nil {
				defaultValue = value
			} else if _, e := strconv.ParseBool(value); e == nil {
				defaultValue = value
			} else if _, e := strconv.ParseFloat(value, 64); e == nil {
				defaultValue = value
			} else if strings.Contains(value, "()") {
				defaultValue = value
			}

			defaultValueClause = fmt.Sprintf("DEFAULT %s", defaultValue)
		}

	}

	isNullableClause := "NULL"
	if !column.IsNullable {
		isNullableClause = "NOT NULL"
	}

	isUniqueClause := ""
	if column.IsUnique {
		isUniqueClause = "UNIQUE"
	}

	q := fmt.Sprintf("%s %s %s %s %s", column.Name, column.DataType, defaultValueClause, isNullableClause, isUniqueClause)
	return q, nil
}

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
	sql, err := getFkQuery(objects.UpdateRelationCreate, relation)
	if err != nil {
		return err
	}

	logger.Debug("Create foreign key - execute : ", sql)
	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("create foreign key %s.%s error : %s", relation.SourceTableName, relation.SourceColumnName, err)
	}

	return nil
}

func updateForeignKey(cfg *raiden.Config, relation *objects.TablesRelationship) error {
	deleteSql, err := getFkQuery(objects.UpdateRelationDelete, relation)
	if err != nil {
		return err
	}

	createSql, err := getFkQuery(objects.UpdateRelationCreate, relation)
	if err != nil {
		return err
	}

	sql := deleteSql + createSql
	logger.Debug("Update foreign key - execute : ", sql)
	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("update foreign key %s.%s error : %s", relation.SourceTableName, relation.SourceColumnName, err)
	}

	return nil
}

func deleteForeignKey(cfg *raiden.Config, relation *objects.TablesRelationship) error {
	sql, err := getFkQuery(objects.UpdateRelationDelete, relation)
	if err != nil {
		return err
	}
	logger.Debug("Delete foreign key - execute : ", sql)
	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("delete foreign key %s.%s error : %s", relation.SourceTableName, relation.SourceColumnName, err)
	}
	return nil
}

func getFkQuery(updateType objects.UpdateRelationType, relation *objects.TablesRelationship) (string, error) {
	alter := fmt.Sprintf("ALTER TABLE IF EXISTS %s.%s", relation.SourceSchema, relation.SourceTableName)
	switch updateType {
	case objects.UpdateRelationCreate:
		return fmt.Sprintf("%s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s.%s (%s);",
			alter, relation.ConstraintName, relation.SourceColumnName,
			relation.TargetTableSchema, relation.TargetTableName, relation.TargetColumnName,
		), nil
	case objects.UpdateRelationDelete:
		return fmt.Sprintf("%s DROP CONSTRAINT IF EXISTS %s;", alter, relation.ConstraintName), nil
	default:
		return "", fmt.Errorf("update relation with type '%s' is not available", updateType)
	}
}
