package state

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"sync"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
)

type (
	CompareDiffType     string
	CompareDiffCategory string

	CompareDiffResult struct {
		Name             string
		Category         CompareDiffCategory
		SupabaseResource any
		AppResource      any
	}
)

var (
	CompareDiffCategoryConflict      CompareDiffCategory = "conflict"
	CompareDiffCategoryCloudNofFound CompareDiffCategory = "cloud-not-found"
	CompareDiffCategoryAppNotFound   CompareDiffCategory = "app-not-found"
)

func ToSupabaseTable(tableStates []TableState) (tables []supabase.Table, err error) {
	fset, astFiles, err := loadAllFile(tableStates)
	if err != nil {
		return tables, err
	}

	conf := types.Config{}
	conf.Importer = importer.Default()
	pkg, err := conf.Check("models", fset, astFiles, nil)
	if err != nil {
		fmt.Println("Error type-checking code:", err)
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

func loadAllFile(tableStates []TableState) (mainFset *token.FileSet, astFiles []*ast.File, err error) {
	type loadResult struct {
		Ast        *ast.File
		Err        error
		Fset       *token.FileSet
		StructName string
	}

	loadChan := make(chan *loadResult)

	wg := sync.WaitGroup{}
	wg.Add(len(tableStates))

	go func() {
		wg.Wait()
		close(loadChan)
	}()

	for i := range tableStates {
		s := tableStates[i]
		go func(w *sync.WaitGroup, state *TableState, lChan chan *loadResult) {
			defer w.Done()
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, state.ModelPath, nil, parser.ParseComments)
			if err != nil {
				lChan <- &loadResult{Err: err}
			} else {
				lChan <- &loadResult{Ast: file, StructName: state.ModelStruct, Fset: fset, Err: nil}
			}
		}(&wg, &s, loadChan)
	}

	for rs := range loadChan {
		if rs.Err != nil {
			return nil, nil, err
		}
		if mainFset == nil {
			mainFset = rs.Fset
		}

		astFiles = append(astFiles, rs.Ast)
	}

	return
}

func createTableFromState(pkg *types.Package, astFiles []*ast.File, fset *token.FileSet, state TableState) (table supabase.Table, err error) {
	obj := pkg.Scope().Lookup(state.ModelStruct)
	if obj == nil {
		fmt.Println("Struct not found:", state.ModelStruct)
		return
	}

	// Assert the object's type to *types.TypeName
	typeObj, ok := obj.(*types.TypeName)
	if !ok {
		fmt.Println("Unexpected type for object:", obj)
		return
	}

	// Get the reflect.Type of the struct
	structType := typeObj.Type().Underlying().(*types.Struct)
	table = state.Table
	table.Name = utils.ToSnakeCase(typeObj.Name())

	// map column for make check if column exist and reuse default
	mapColumn := make(map[string]supabase.Column)
	for i := range table.Columns {
		c := table.Columns[i]
		mapColumn[c.Name] = c
	}

	// map relation for make check if relation exist and reuse default
	mapRelation := make(map[string]supabase.TablesRelationship)
	for i := range table.Relationships {
		r := table.Relationships[i]
		mapRelation[r.ConstraintName] = r
	}

	// map relation for make check if relation exist and reuse default
	mapPrimaryKey := make(map[string]supabase.PrimaryKey)
	for i := range table.PrimaryKeys {
		pk := table.PrimaryKeys[i]
		mapPrimaryKey[pk.Name] = pk
	}

	var tableColumns []supabase.Column
	var tableRealtions []supabase.TablesRelationship
	var tablePrimaryKeys []supabase.PrimaryKey

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
						tablePrimaryKeys = append(tablePrimaryKeys, supabase.PrimaryKey{
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

func buildColumn(fieldName string, mapColumn map[string]supabase.Column, columnTag string) (column supabase.Column) {
	ct := raiden.MarshalColumnTag(columnTag)

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

func buildRelation(tableName, fieldName, schema string, mapRelations map[string]supabase.TablesRelationship, joinTag string) (relation supabase.TablesRelationship) {
	jt := raiden.MarshallJoinTag(joinTag)
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

func CompareTables(supabaseTable []supabase.Table, appTable []supabase.Table) (diffResult []CompareDiffResult, err error) {
	mapAppTable := make(map[int]supabase.Table)
	for i := range appTable {
		t := appTable[i]
		logger.Debug("[pkg.state.table] app table id : ", t.ID)
		mapAppTable[t.ID] = t
	}

	for i := range supabaseTable {
		t := supabaseTable[i]

		appTable, isExist := mapAppTable[t.ID]
		logger.Debug("[pkg.state.table] supabase check id : ", t.ID, " is exist : ", isExist)
		if isExist {
			spByte, err := json.Marshal(t)
			if err != nil {
				return diffResult, err
			}
			spHash := utils.HashByte(spByte)

			appByte, err := json.Marshal(appTable)
			if err != nil {
				return diffResult, err
			}
			appHash := utils.HashByte(appByte)

			if spHash != appHash {
				diffResult = append(diffResult, CompareDiffResult{
					Name:             t.Name,
					Category:         CompareDiffCategoryConflict,
					SupabaseResource: t,
					AppResource:      appTable,
				})
			}
		}
	}

	return
}
