package query

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func BuildCreateTableQuery(newTable objects.Table) (string, error) {
	createSql, err := buildCreateTableQuery(newTable.Schema, newTable)
	if err != nil {
		return "", err
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
	return sql, nil
}

func BuildUpdateTableQuery(newTable objects.Table, updateItem objects.UpdateTableParam) string {
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

	return sql
}

func BuildDeleteTableQuery(table objects.Table, cascade bool) string {
	sql := fmt.Sprintf("DROP TABLE %s.%s", table.Schema, table.Name)
	if cascade {
		sql += " CASCADE"
	} else {
		sql += " RESTRICT"
	}
	sql += ";"
	return sql
}

func buildCreateTableQuery(schema string, table objects.Table) (q string, err error) {
	var tableContains []string

	// add column definition
	for i := range table.Columns {
		c := table.Columns[i]
		colDef, err := buildColumnDef(c)
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

	q = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (%s);", schema, table.Name, strings.Join(tableContains, ","))
	return
}

// ----- Column ----
func BuildCreateColumnQuery(column objects.Column, isPrimary bool) (q string, err error) {
	colDef, err := buildColumnDef(column)
	if err != nil {
		return q, err
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

	q = fmt.Sprintf(`
	BEGIN;
	  ALTER TABLE %s.%s ADD COLUMN %s %s;
	COMMIT;`, column.Schema, column.Table, colDef, isPrimaryKeyClause)
	return
}

func BuildUpdateColumnQuery(oldColumn, newColumn objects.Column, updateItem objects.UpdateColumnItem) (q string) {
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
	q = "BEGIN;"
	for _, stmt := range sqlStatements {
		q += " " + stmt
	}
	q += " COMMIT;"

	return q
}

func buildColumnDef(column objects.Column) (string, error) {
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

func BuildDeleteColumnQuery(column objects.Column) (q string) {
	return fmt.Sprintf("ALTER TABLE %s.%s DROP COLUMN %s;", column.Schema, column.Table, column.Name)
}

func BuildFkQuery(updateType objects.UpdateRelationType, relation *objects.TablesRelationship) (string, error) {
	alter := fmt.Sprintf("ALTER TABLE IF EXISTS %s.%s", relation.SourceSchema, relation.SourceTableName)
	switch updateType {
	case objects.UpdateRelationCreate:
		tmp := `
		do $$
		BEGIN
			IF NOT EXISTS (SELECT CONSTRAINT_NAME FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS WHERE CONSTRAINT_NAME = '%s' AND TABLE_NAME = '%s') THEN
				%s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s.%s (%s) %s %s;
			END IF;
		END $$;
		`

		var onUpdate, onDelete string

		if relation.Action != nil {
			if relation.Action.UpdateAction != "" {
				action := relation.Action.UpdateAction
				if len(action) == 1 {
					action = string(objects.RelationActionMapLabel[objects.RelationAction(action)])
				}

				onUpdate = fmt.Sprintf("ON UPDATE %s", strings.ToUpper(action))
			}

			if relation.Action.DeletionAction != "" {
				action := relation.Action.DeletionAction
				if len(action) == 1 {
					action = string(objects.RelationActionMapLabel[objects.RelationAction(action)])
				}

				onDelete = fmt.Sprintf("ON DELETE %s", strings.ToUpper(action))
			}
		}

		return fmt.Sprintf(tmp, relation.ConstraintName, relation.SourceTableName,
			alter, relation.ConstraintName, relation.SourceColumnName,
			relation.TargetTableSchema, relation.TargetTableName, relation.TargetColumnName, onUpdate, onDelete,
		), nil
	case objects.UpdateRelationDelete:
		return fmt.Sprintf("%s DROP CONSTRAINT IF EXISTS %s;", alter, relation.ConstraintName), nil
	default:
		return "", fmt.Errorf("update relation with type '%s' is not available", updateType)
	}
}

func BuildFKIndexQuery(updateType objects.UpdateRelationType, relation *objects.TablesRelationship) (string, error) {
	if relation == nil {
		return "", nil
	}

	indexName := fmt.Sprintf("ix_%s_%s", relation.SourceTableName, relation.SourceColumnName)

	switch updateType {
	case objects.UpdateRelationCreate:
		tmp := `
		DO $$
		BEGIN
			-- Check if the index already exists
			IF NOT EXISTS (
				SELECT 1 
				FROM pg_class c
				JOIN pg_namespace n ON n.oid = c.relnamespace
				WHERE c.relname = '%s'  -- Replace with your index name
				AND n.nspname = '%s'  -- Replace with the schema name if necessary
			) THEN
				-- Create the index if it does not exist
				CREATE INDEX %s ON %s.%s (%s);
			END IF;
		END $$;
		`
		return fmt.Sprintf(tmp, indexName, relation.SourceSchema, indexName, relation.SourceSchema, relation.SourceTableName, relation.SourceColumnName), nil
	case objects.UpdateRelationDelete:
		tmp := `
		DO $$
		BEGIN
			-- Check if the index already exists
			IF NOT EXISTS (
				SELECT 1 
				FROM pg_class c
				JOIN pg_namespace n ON n.oid = c.relnamespace
				WHERE c.relname = '%s'  -- Replace with your index name
				AND n.nspname = '%s'  -- Replace with the schema name if necessary
			) THEN
			 	-- Drop the index if it exists
        		EXECUTE 'DROP INDEX %s.%s';  -- Ensure to specify the correct schema
			END IF;
		END $$;
		`
		return fmt.Sprintf(tmp, indexName, relation.SourceSchema, relation.SourceSchema, indexName), nil
	default:
		return "", fmt.Errorf("update index with type '%s' is not available", updateType)
	}
}
