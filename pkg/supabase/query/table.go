package query

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/lib/pq"
	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func BuildCreateTableQuery(newTable objects.Table) (string, error) {
	createSql, err := buildCreateTableQuery(newTable.Schema, newTable)
	if err != nil {
		return "", err
	}
	schemaIdent := pq.QuoteIdentifier(newTable.Schema)
	tableIdent := pq.QuoteIdentifier(newTable.Name)

	var rlsEnableQuery string
	if newTable.RLSEnabled {
		rlsEnableQuery = fmt.Sprintf("ALTER TABLE %s.%s ENABLE ROW LEVEL SECURITY;", schemaIdent, tableIdent)
	}

	var rlsForcedQuery string
	if newTable.RLSForced {
		rlsForcedQuery = fmt.Sprintf("ALTER TABLE %s.%s FORCE ROW LEVEL SECURITY;", schemaIdent, tableIdent)
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
	alter := fmt.Sprintf("ALTER TABLE %s.%s", pq.QuoteIdentifier(updateItem.OldData.Schema), pq.QuoteIdentifier(updateItem.OldData.Name))
	for _, uType := range updateItem.ChangeItems {
		switch uType {
		case objects.UpdateTableSchema:
			schemaQuery = fmt.Sprintf("%s SET SCHEMA %s;", alter, pq.QuoteIdentifier(newTable.Schema))
		case objects.UpdateTableName:
			if newTable.Name != "" {
				nameQuery = fmt.Sprintf("%s RENAME TO %s;", alter, pq.QuoteIdentifier(newTable.Name))
			}
		case objects.UpdateTableRlsEnable:
			if newTable.RLSEnabled {
				enableRlsQuery = fmt.Sprintf("%s ENABLE ROW LEVEL SECURITY;", alter)
			} else {
				enableRlsQuery = fmt.Sprintf("%s DISABLE ROW LEVEL SECURITY;", alter)
			}
		case objects.UpdateTableRlsForced:
			if newTable.RLSForced {
				forceRlsQuery = fmt.Sprintf("%s FORCE ROW LEVEL SECURITY;", alter)
			} else {
				forceRlsQuery = fmt.Sprintf("%s NO FORCE ROW LEVEL SECURITY;", alter)
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
					pkArr = append(pkArr, pq.QuoteIdentifier(v.Name))
				}
				primaryKeysQuery += fmt.Sprintf("%s ADD PRIMARY KEY (%s);", alter, strings.Join(pkArr, ", "))
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
	schemaIdent := pq.QuoteIdentifier(table.Schema)
	tableIdent := pq.QuoteIdentifier(table.Name)
	sql := fmt.Sprintf("DROP TABLE %s.%s", schemaIdent, tableIdent)
	if cascade {
		sql += " CASCADE"
	} else {
		sql += " RESTRICT"
	}
	sql += ";"
	return sql
}

func buildCreateTableQuery(schema string, table objects.Table) (q string, err error) {
	schemaIdent := pq.QuoteIdentifier(schema)
	tableIdent := pq.QuoteIdentifier(table.Name)
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
		primaryKeys = append(primaryKeys, pq.QuoteIdentifier(pk.Name))
	}

	if len(primaryKeys) > 0 {
		tableContains = append(tableContains, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(primaryKeys, ",")))
	}

	q = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (%s);", schemaIdent, tableIdent, strings.Join(tableContains, ","))
	return
}

// ----- Column ----
func BuildCreateColumnQuery(column objects.Column, isPrimary bool) (q string, err error) {
	colDef, err := buildColumnDef(column)
	if err != nil {
		return q, err
	}

	schemaIdent := pq.QuoteIdentifier(column.Schema)
	tableIdent := pq.QuoteIdentifier(column.Table)
	statement := fmt.Sprintf("ALTER TABLE %s.%s ADD COLUMN %s", schemaIdent, tableIdent, colDef)
	if isPrimary {
		statement += " PRIMARY KEY"
	}
	statement += ";"

	q = fmt.Sprintf("BEGIN; %s COMMIT;", statement)
	return
}

func BuildUpdateColumnQuery(oldColumn, newColumn objects.Column, updateItem objects.UpdateColumnItem) (q string) {
	// Prepare SQL statements
	var sqlStatements []string
	schemaIdent := pq.QuoteIdentifier(newColumn.Schema)
	tableIdent := pq.QuoteIdentifier(newColumn.Table)
	var alter = fmt.Sprintf("ALTER TABLE %s.%s", schemaIdent, tableIdent)
	currentColumnIdent := pq.QuoteIdentifier(oldColumn.Name)
	for _, uType := range updateItem.UpdateItems {
		switch uType {
		case objects.UpdateColumnName:
			if newColumn.Name != "" {
				newColumnIdent := pq.QuoteIdentifier(newColumn.Name)
				sqlStatements = append(
					sqlStatements,
					fmt.Sprintf(
						"%s RENAME COLUMN %s TO %s;", alter, currentColumnIdent, newColumnIdent,
					),
				)
				currentColumnIdent = newColumnIdent
			}
		case objects.UpdateColumnDataType:
			dataType := newColumn.DataType

			if dataType == string(postgres.UserDefined) {
				dataType = newColumn.Format
			}

			sqlStatements = append(
				sqlStatements,
				fmt.Sprintf(
					"%s ALTER COLUMN %s SET DATA TYPE %s USING %s::%s;", alter, currentColumnIdent, dataType, currentColumnIdent, dataType,
				),
			)
		case objects.UpdateColumnUnique:
			constraintIdent := pq.QuoteIdentifier(fmt.Sprintf("%s_%s_unique", newColumn.Table, newColumn.Name))
			if newColumn.IsUnique {
				sqlStatements = append(
					sqlStatements,
					fmt.Sprintf(
						"%s ADD CONSTRAINT %s UNIQUE (%s);", alter, constraintIdent, pq.QuoteIdentifier(newColumn.Name)),
				)
			} else {
				sqlStatements = append(
					sqlStatements,
					fmt.Sprintf(
						"%s DROP CONSTRAINT %s;", alter, constraintIdent,
					),
				)
			}
		case objects.UpdateColumnNullable:
			if newColumn.IsNullable {
				sqlStatements = append(
					sqlStatements,
					fmt.Sprintf(
						"%s ALTER COLUMN %s DROP NOT NULL;", alter, currentColumnIdent,
					),
				)
			} else {
				sqlStatements = append(
					sqlStatements,
					fmt.Sprintf(
						"%s ALTER COLUMN %s SET NOT NULL;", alter, currentColumnIdent,
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

			defaultValue := pq.QuoteLiteral(value)
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
					"%s ALTER COLUMN %s SET DEFAULT %s;", alter, currentColumnIdent, defaultValue,
				),
			)

		case objects.UpdateColumnIdentity:
			if newColumn.IsIdentity {
				sqlStatements = append(
					sqlStatements,
					fmt.Sprintf(
						"%s ALTER COLUMN %s ADD GENERATED %s AS IDENTITY;", alter, currentColumnIdent, newColumn.IdentityGeneration,
					),
				)
			} else {
				sqlStatements = append(
					sqlStatements,
					fmt.Sprintf(
						"%s ALTER COLUMN %s DROP IDENTITY IF EXISTS;", alter, currentColumnIdent,
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

			// handle reserved default value keyword
			if _, exist := sql.MapDefaultFunctionValue[value]; exist {
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

	dataType := column.DataType
	if dataType == string(postgres.UserDefined) {
		dataType = column.Format
	}

	q := strings.TrimSpace(fmt.Sprintf("%s %s %s %s %s", pq.QuoteIdentifier(column.Name), dataType, defaultValueClause, isNullableClause, isUniqueClause))
	return q, nil
}

func BuildDeleteColumnQuery(column objects.Column) (q string) {
	return fmt.Sprintf("ALTER TABLE %s.%s DROP COLUMN %s;", pq.QuoteIdentifier(column.Schema), pq.QuoteIdentifier(column.Table), pq.QuoteIdentifier(column.Name))
}

func BuildFkQuery(updateType objects.UpdateRelationType, relation *objects.TablesRelationship) (string, error) {
	alter := fmt.Sprintf("ALTER TABLE IF EXISTS %s.%s", pq.QuoteIdentifier(relation.SourceSchema), pq.QuoteIdentifier(relation.SourceTableName))
	switch updateType {
	case objects.UpdateRelationCreate:
		sourceColumn := pq.QuoteIdentifier(relation.SourceColumnName)
		targetSchema := pq.QuoteIdentifier(relation.TargetTableSchema)
		targetTable := pq.QuoteIdentifier(relation.TargetTableName)
		targetColumn := pq.QuoteIdentifier(relation.TargetColumnName)
		constraintIdent := pq.QuoteIdentifier(relation.ConstraintName)
		constraintLiteral := pq.QuoteLiteral(relation.ConstraintName)
		sourceTableLiteral := pq.QuoteLiteral(relation.SourceTableName)
		sourceSchemaLiteral := pq.QuoteLiteral(relation.SourceSchema)

		var onUpdate, onDelete string

		if relation.Action != nil {
			if relation.Action.UpdateAction != "" {
				action := relation.Action.UpdateAction
				if len(action) == 1 {
					action = string(objects.RelationActionMapLabel[objects.RelationAction(action)])
				}

				onUpdate = fmt.Sprintf(" ON UPDATE %s", strings.ToUpper(action))
			}

			if relation.Action.DeletionAction != "" {
				action := relation.Action.DeletionAction
				if len(action) == 1 {
					action = string(objects.RelationActionMapLabel[objects.RelationAction(action)])
				}

				onDelete = fmt.Sprintf(" ON DELETE %s", strings.ToUpper(action))
			}
		}

		return fmt.Sprintf(`DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1 FROM information_schema.table_constraints
		WHERE constraint_name = %s
		  AND table_schema = %s
		  AND table_name = %s
	) THEN
		%s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s.%s (%s)%s%s;
	END IF;
END $$;`,
			constraintLiteral,
			sourceSchemaLiteral,
			sourceTableLiteral,
			alter,
			constraintIdent,
			sourceColumn,
			targetSchema,
			targetTable,
			targetColumn,
			onUpdate,
			onDelete,
		), nil
	case objects.UpdateRelationDelete:
		return fmt.Sprintf("%s DROP CONSTRAINT IF EXISTS %s;", alter, pq.QuoteIdentifier(relation.ConstraintName)), nil
	default:
		return "", fmt.Errorf("update relation with type '%s' is not available", updateType)
	}
}

func BuildFKIndexQuery(updateType objects.UpdateRelationType, relation *objects.TablesRelationship) (string, error) {
	if relation == nil {
		return "", nil
	}

	indexName := fmt.Sprintf("ix_%s_%s", relation.SourceTableName, relation.SourceColumnName)
	indexIdent := pq.QuoteIdentifier(indexName)
	schemaIdent := pq.QuoteIdentifier(relation.SourceSchema)
	tableIdent := pq.QuoteIdentifier(relation.SourceTableName)
	columnIdent := pq.QuoteIdentifier(relation.SourceColumnName)

	switch updateType {
	case objects.UpdateRelationCreate:
		return fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s.%s (%s);", indexIdent, schemaIdent, tableIdent, columnIdent), nil
	case objects.UpdateRelationDelete:
		return fmt.Sprintf("DROP INDEX IF EXISTS %s.%s;", schemaIdent, indexIdent), nil
	default:
		return "", fmt.Errorf("update index with type '%s' is not available", updateType)
	}
}
