package state

import (
	"errors"
	"fmt"
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

func ToTables(tableStates []TableState) (tables []objects.Table, err error) {
	var paths []string
	for i := range tableStates {
		t := tableStates[i]
		paths = append(paths, t.ModelPath)
	}

	fset, astFiles, err := loadFiles(paths)
	if err != nil {
		return tables, err
	}

	conf := types.Config{}
	conf.Importer = importer.Default()
	pkg, err := conf.Check("models", fset, astFiles, nil)
	if err != nil {
		err = fmt.Errorf("error type-checking code :  %s", err.Error())
		return
	}

	for i := range tableStates {
		t := tableStates[i]
		tb, err := createTableFromState(pkg, astFiles, fset, t)
		if err != nil {
			return tables, err
		}
		tables = append(tables, tb)
	}

	return
}

func createTableFromState(pkg *types.Package, astFiles []*ast.File, fset *token.FileSet, state TableState) (table objects.Table, err error) {
	obj := pkg.Scope().Lookup(state.ModelStruct)
	if obj == nil {
		err = errors.New("struct not found : " + state.ModelStruct)
		return
	}

	// Assert the objects's type to *types.TypeName
	typeObj, ok := obj.(*types.TypeName)
	if !ok {
		err = fmt.Errorf("unexpected type for objects : %v", obj)
		return
	}

	// Get the reflect.Type of the struct
	structType := typeObj.Type().Underlying().(*types.Struct)
	table = state.Table
	table.Name = utils.ToSnakeCase(typeObj.Name())

	// map column for make check if column exist and reuse default
	mapColumn := make(map[string]objects.Column)
	for i := range table.Columns {
		c := table.Columns[i]
		mapColumn[c.Name] = c
	}

	// map relation for make check if relation exist and reuse default
	mapRelation := make(map[string]objects.TablesRelationship)
	for i := range table.Relationships {
		r := table.Relationships[i]
		mapRelation[r.ConstraintName] = r
	}

	// map relation for make check if relation exist and reuse default
	mapPrimaryKey := make(map[string]objects.PrimaryKey)
	for i := range table.PrimaryKeys {
		pk := table.PrimaryKeys[i]
		mapPrimaryKey[pk.Name] = pk
	}

	var tableColumns []objects.Column
	var tableRealtions []objects.TablesRelationship
	var tablePrimaryKeys []objects.PrimaryKey

	// Iterate over the fields of the struct
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		// Get field name and tag
		fieldName := field.Name()

		// example tag :
		// "column": "name:id;type:bigint;primaryKey;autoIncrement;nullable:false"
		// "join": "joinType:manyToMany;through:submission;sourcePrimaryKey:id;sourceForeignKey:scouter_id;targetPrimaryKey:id;targetForeign:scouter_id"
		fieldTag := structType.Tag(i)

		// change tag to map
		mapTag := make(map[string]string)
		for _, rawTag := range strings.Split(fieldTag, " ") {
			parts := strings.SplitN(rawTag, ":", 2)
			if len(parts) == 2 {
				key := strings.Trim(parts[0], `"`)
				value := strings.Trim(parts[1], `"`)
				mapTag[key] = value
			}
		}

		if fieldName == "Metadata" {
			if schema, isSet := mapTag["schema"]; isSet {
				table.Schema = schema
			}
		}

		if fieldName == "Acl" {
			continue
		}

		columnTag, exist := mapTag["column"]
		if exist {
			if c := buildColumn(fieldName, mapColumn, columnTag); c.Name != "" {
				tableColumns = append(tableColumns, c)
				if c.IsIdentity {
					if pk, exist := mapPrimaryKey[c.Name]; exist {
						pk.Schema = c.Schema
						pk.TableName = table.Name
						tablePrimaryKeys = append(tablePrimaryKeys, pk)
					} else {
						tablePrimaryKeys = append(tablePrimaryKeys, objects.PrimaryKey{
							Name:      c.Name,
							Schema:    table.Schema,
							TableID:   table.ID,
							TableName: table.Name,
						})
					}
				}
			}
		}

		joinTag, exist := mapTag["join"]
		if exist {
			if r := buildRelation(table.Name, fieldName, table.Schema, mapRelation, joinTag); r.ConstraintName != "" {
				tableRealtions = append(tableRealtions, r)
			}
		}
	}

	table.Columns = tableColumns
	table.Relationships = tableRealtions
	table.PrimaryKeys = tablePrimaryKeys

	return table, nil
}

func buildColumn(fieldName string, mapColumn map[string]objects.Column, columnTag string) (column objects.Column) {
	ct := raiden.UnmarshalColumnTag(columnTag)

	if ct.Name == "" {
		ct.Name = utils.ToSnakeCase(fieldName)
	}

	if c, exist := mapColumn[ct.Name]; exist {
		column = c
	}

	if ct.PrimaryKey {
		column.IsIdentity = ct.PrimaryKey
	}

	column.IsNullable = ct.Nullable
	column.DataType = string(postgres.GetPgDataTypeName(postgres.DataType(ct.Type), false))
	column.DefaultValue = ct.Default
	column.IsUnique = ct.Unique

	if ct.AutoIncrement {
		column.IdentityGeneration = "BY DEFAULT"
	}

	if len(column.Enums) == 0 {
		column.Enums = make([]string, 0)
	}

	return
}

func buildRelation(tableName, fieldName, schema string, mapRelations map[string]objects.TablesRelationship, joinTag string) (relation objects.TablesRelationship) {
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
		foreignKey = fmt.Sprintf("%s_id", utils.ToSnakeCase(targetTableName))
	} else {
		foreignKey = jt.ForeignKey
	}

	if jt.PrimaryKey != "" {
		primaryKey = "id"
	} else {
		primaryKey = jt.PrimaryKey
	}

	// overwrite with default if relation is exist
	relation.ConstraintName = getRelationConstrainName(sourceTableName, foreignKey)
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
func getRelationConstrainName(sourceTable, sourceColumn string) string {
	return fmt.Sprintf("%s_%s_fkey", sourceTable, sourceColumn)
}
