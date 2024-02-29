package state

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type ExtractTableItem struct {
	Table    objects.Table
	Policies objects.Policies
}

type ExtractTableResult struct {
	Existing []ExtractTableItem
	New      []ExtractTableItem
	Delete   []ExtractTableItem
}

func ExtractTable(tableStates []TableState, appTable []any) (result ExtractTableResult, err error) {
	var mapTableState = make(map[string]TableState)

	for i := range tableStates {
		t := tableStates[i]

		mapTableState[t.Table.Name] = t
	}

	for _, t := range appTable {
		tableType := reflect.TypeOf(t)
		if tableType.Kind() == reflect.Ptr {
			tableType = tableType.Elem()
		}

		tableName := utils.ToSnakeCase(tableType.Name())
		ts, isExist := mapTableState[tableName]
		logger.Debug("check table : ", tableName, " is exist : ", isExist)
		if !isExist {
			nt := buildTableFromModel(t)
			result.New = append(result.New, nt)
			continue
		}

		tb, e := buildTableFromState(t, ts)
		if e != nil {
			err = e
			return
		}
		result.Existing = append(result.Existing, tb)

		delete(mapTableState, tableName)
	}

	// if table from state is not exist in app table,
	// add table to deleted table arr
	for _, v := range mapTableState {
		result.Delete = append(result.Delete, ExtractTableItem{
			Table: v.Table,
		})
	}

	return
}

func buildTableFromModel(model any) (ei ExtractTableItem) {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	ei.Table.Name = utils.ToSnakeCase(modelType.Name())

	// add metadata
	metadataField, isExist := modelType.FieldByName("Metadata")
	if isExist {
		bindTableMetadata(&metadataField, &ei.Table)
	}

	// add metadata
	aclField, isExist := modelType.FieldByName("Acl")
	if isExist {
		bindTableAcl(&aclField)
	}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		switch field.Name {
		case "Metadata":
			continue
		case "Acl":
			// TODO : Implement this
			continue
		default:
			if column := field.Tag.Get("column"); len(column) > 0 {
				ct := raiden.UnmarshalColumnTag(column)

				c := objects.Column{
					Table:  ei.Table.Name,
					Schema: ei.Table.Schema,
				}

				bindColumn(&field, &ct, &c)
				ei.Table.Columns = append(ei.Table.Columns, c)

				if ct.PrimaryKey {
					ei.Table.PrimaryKeys = append(ei.Table.PrimaryKeys, objects.PrimaryKey{
						Name:      c.Name,
						TableName: c.Table,
						Schema:    c.Schema,
					})
				}
			}

			if join := field.Tag.Get("join"); len(join) > 0 {
				jt := raiden.UnmarshalJoinTag(join)
				if jt.JoinType == raiden.RelationTypeHasOne {
					rel := objects.TablesRelationship{}
					rel.SourceTableName = ei.Table.Name
					rel.SourceColumnName = jt.ForeignKey
					rel.SourceSchema = ei.Table.Schema
					rel.TargetTableName = utils.ToSnakeCase(field.Name)
					rel.TargetTableSchema = ei.Table.Schema
					rel.TargetColumnName = jt.PrimaryKey

					ei.Table.Relationships = append(ei.Table.Relationships, rel)
				}
			}
		}
	}

	return
}

func buildTableFromState(model any, state TableState) (ei ExtractTableItem, err error) {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// Get the reflect.Type of the struct
	ei.Table = state.Table
	ei.Table.Name = utils.ToSnakeCase(modelType.Name())

	// map column for make check if column exist and reuse default
	mapColumn := make(map[string]objects.Column)
	for i := range ei.Table.Columns {
		c := ei.Table.Columns[i]
		mapColumn[c.Name] = c
	}

	// map relation for make check if relation exist and reuse default
	mapRelation := make(map[string]objects.TablesRelationship)
	for i := range ei.Table.Relationships {
		r := ei.Table.Relationships[i]
		mapRelation[r.ConstraintName] = r
	}

	// map relation for make check if relation exist and reuse default
	mapPrimaryKey := make(map[string]objects.PrimaryKey)
	for i := range ei.Table.PrimaryKeys {
		pk := ei.Table.PrimaryKeys[i]
		mapPrimaryKey[pk.Name] = pk
	}

	var columns []objects.Column
	var relations []objects.TablesRelationship
	var primaryKeys []objects.PrimaryKey

	// update metadata
	metadataField, isExist := modelType.FieldByName("Metadata")
	if isExist {
		bindTableMetadata(&metadataField, &ei.Table)
	}

	// add metadata
	aclField, isExist := modelType.FieldByName("Acl")
	if isExist {
		bindTableAcl(&aclField)
	}

	// Iterate over the fields of the struct
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		// Get field name and tag
		fieldName := field.Name

		switch field.Name {
		case "Metadata", "Acl":
			continue
		default:
			// example tag "name:id;type:bigint;primaryKey;autoIncrement;nullable:false"
			if columnTag := field.Tag.Get("column"); len(columnTag) > 0 {
				var c objects.Column

				ct := raiden.UnmarshalColumnTag(columnTag)
				if found, exist := mapColumn[ct.Name]; exist {
					c = found
				}

				c.Table = ei.Table.Name
				c.Schema = ei.Table.Schema

				bindColumn(&field, &ct, &c)

				if c.IsIdentity {
					if pk, exist := mapPrimaryKey[c.Name]; exist {
						pk.Schema = c.Schema
						pk.TableName = ei.Table.Name
						primaryKeys = append(primaryKeys, pk)
					} else {
						primaryKeys = append(primaryKeys, objects.PrimaryKey{
							Name:      c.Name,
							Schema:    ei.Table.Schema,
							TableID:   ei.Table.ID,
							TableName: ei.Table.Name,
						})
					}
				}

				columns = append(columns, c)
			}

			if joinTag := field.Tag.Get("join"); len(joinTag) > 0 {
				if r := buildTableRelation(ei.Table.Name, fieldName, ei.Table.Schema, mapRelation, joinTag); r.ConstraintName != "" {
					relations = append(relations, r)
				}
			}
		}
	}

	ei.Table.Columns = columns
	ei.Table.Relationships = relations
	ei.Table.PrimaryKeys = primaryKeys

	return ei, nil
}

func bindColumn(field *reflect.StructField, ct *raiden.ColumnTag, c *objects.Column) {
	c.IsNullable = ct.Nullable
	c.IsUnique = ct.Unique

	if ct.Name != "" {
		c.Name = ct.Name
	} else {
		ct.Name = utils.ToSnakeCase(field.Name)
	}

	if ct.PrimaryKey {
		c.IsIdentity = true
	}

	if ct.Type != "" {
		pgType := postgres.GetPgDataTypeName(postgres.DataType(ct.Type), false)
		c.DataType = string(pgType)
	} else {
		c.DataType = string(postgres.ToPostgresType(field.Type.Name()))
	}

	c.DefaultValue = ct.Default

	if ct.AutoIncrement {
		c.IdentityGeneration = "BY DEFAULT"
	}

	if len(c.Enums) == 0 {
		c.Enums = make([]string, 0)
	}
}

func bindTableMetadata(field *reflect.StructField, table *objects.Table) {
	if schema := field.Tag.Get("schema"); len(schema) > 0 {
		table.Schema = schema
	} else {
		table.Schema = "public"
	}

	if rlsEnable := field.Tag.Get("rlsEnable"); len(rlsEnable) > 0 {
		if isRlsEnable, err := strconv.ParseBool(rlsEnable); err == nil {
			table.RLSEnabled = isRlsEnable
		}
	} else {
		table.RLSEnabled = false
	}

	if rlsForced := field.Tag.Get("rlsForced"); len(rlsForced) > 0 {
		if isRlsForced, err := strconv.ParseBool(rlsForced); err == nil {
			table.RLSForced = isRlsForced
		}
	} else {
		table.RLSForced = false
	}
}

// TODO : implement mapping acl
func bindTableAcl(field *reflect.StructField) {

}

func buildTableRelation(tableName, fieldName, schema string, mapRelations map[string]objects.TablesRelationship, joinTag string) (relation objects.TablesRelationship) {
	jt := raiden.UnmarshalJoinTag(joinTag)

	sourceTable, targetTable := utils.ToSnakeCase(fieldName), utils.ToSnakeCase(tableName)

	var sourceTableName, targetTableName, primaryKey, foreignKey string

	switch jt.JoinType {
	case raiden.RelationTypeHasMany:
		sourceTableName = sourceTable
		targetTableName = targetTable
	case raiden.RelationTypeHasOne:
		sourceTableName = targetTable
		targetTableName = sourceTable
	case raiden.RelationTypeManyToMany:
		return
	default:
		return
	}

	// setup primary and foreign key
	if jt.ForeignKey != "" {
		foreignKey = jt.ForeignKey
	} else {
		foreignKey = fmt.Sprintf("%s_id", utils.ToSnakeCase(targetTableName))
	}

	if jt.PrimaryKey != "" {
		primaryKey = jt.PrimaryKey
	} else {
		primaryKey = "id"
	}

	// overwrite with default if relation is exist
	relation.ConstraintName = getRelationConstrainName(schema, sourceTableName, foreignKey)
	if r, ok := mapRelations[relation.ConstraintName]; ok {
		relation = r
	}

	relation.SourceSchema = schema
	relation.SourceTableName = sourceTableName
	relation.SourceColumnName = foreignKey

	relation.TargetTableSchema = schema
	relation.TargetTableName = targetTableName
	relation.TargetColumnName = primaryKey

	return
}

// get relation table name, base on struct type that defined in relation field
func getRelationConstrainName(schema, table, foreignKey string) string {
	return fmt.Sprintf("%s_%s_%s_fkey", schema, table, foreignKey)
}
