package state

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type ModelValidationTag map[string]string

type ExtractTableItem struct {
	Table             objects.Table
	ValidationTags    ModelValidationTag
	ExtractedPolicies ExtractPolicyResult
}

type ExtractTableItems []ExtractTableItem

type ExtractTableResult struct {
	Existing ExtractTableItems
	New      ExtractTableItems
	Delete   ExtractTableItems
}

type tableBuilder struct {
	modelType          reflect.Type
	mapDataType        map[string]objects.Type
	hasState           bool
	item               ExtractTableItem
	columns            []objects.Column
	relations          []objects.TablesRelationship
	primaryKeys        []objects.PrimaryKey
	existingColumns    map[string]objects.Column
	existingRelations  map[string]objects.TablesRelationship
	existingPrimaryKey map[string]objects.PrimaryKey
	existingPolicies   map[string]objects.Policy
	acl                *raiden.Acl
}

func ExtractTable(tableStates []TableState, appTable []any, mapDataType map[string]objects.Type) (result ExtractTableResult, err error) {
	var mapTableState = make(map[string]TableState)

	for i := range tableStates {
		t := tableStates[i]
		mapTableState[t.Table.Name] = t
	}

	for _, t := range appTable {
		tableName := raiden.GetTableName(t)
		ts, isExist := mapTableState[tableName]

		if !isExist {
			nt := buildTableFromModel(t, mapDataType)
			result.New = append(result.New, nt)
			continue
		}

		tb := buildTableFromState(t, mapDataType, ts)
		result.Existing = append(result.Existing, tb)

		delete(mapTableState, tableName)
	}

	// if table from state is not exist in app table,
	// add table to deleted table arr
	for _, v := range mapTableState {
		var deletedPolicy []objects.Policy
		if len(v.Policies) > 0 {
			deletedPolicy = append(deletedPolicy, v.Policies...)
		}

		result.Delete = append(result.Delete, ExtractTableItem{
			Table: v.Table,
			ExtractedPolicies: ExtractPolicyResult{
				Delete: deletedPolicy,
			},
		})
	}

	return
}

func buildTableFromModel(model any, mapDataType map[string]objects.Type) (ei ExtractTableItem) {
	b := newTableBuilder(model, mapDataType, nil)
	b.processFields()
	b.collectModelRelations()
	b.applyPolicies()
	return b.finish()
}

func buildTableFromState(model any, mapDataType map[string]objects.Type, state TableState) (ei ExtractTableItem) {
	b := newTableBuilder(model, mapDataType, &state)
	b.processFields()
	b.collectStateRelations()
	b.applyPolicies()
	return b.finish()
}

func newTableBuilder(model any, mapDataType map[string]objects.Type, state *TableState) *tableBuilder {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Pointer {
		modelType = modelType.Elem()
	}

	b := &tableBuilder{
		modelType:          modelType,
		mapDataType:        mapDataType,
		item:               ExtractTableItem{ValidationTags: make(ModelValidationTag)},
		existingColumns:    make(map[string]objects.Column),
		existingRelations:  make(map[string]objects.TablesRelationship),
		existingPrimaryKey: make(map[string]objects.PrimaryKey),
		existingPolicies:   make(map[string]objects.Policy),
	}

	b.item.Table.Name = raiden.GetTableName(model)

	if state != nil {
		b.hasState = true
		b.item.Table = state.Table
		b.item.Table.Name = raiden.GetTableName(model)
		for _, c := range state.Table.Columns {
			b.existingColumns[c.Name] = c
		}
		for _, r := range state.Table.Relationships {
			b.existingRelations[r.ConstraintName] = r
		}
		for _, pk := range state.Table.PrimaryKeys {
			b.existingPrimaryKey[pk.Name] = pk
		}
		for _, p := range state.Policies {
			b.existingPolicies[p.Name] = p
		}
	}

	b.applyMetadata()
	b.columns = make([]objects.Column, 0)
	b.relations = make([]objects.TablesRelationship, 0)
	b.primaryKeys = make([]objects.PrimaryKey, 0)

	// colleact and assign Acl
	b.acl = getAcl(model)

	return b
}

func (b *tableBuilder) applyMetadata() {
	if metadataField, ok := b.modelType.FieldByName("Metadata"); ok {
		bindTableMetadata(&metadataField, &b.item.Table)
		return
	}

	if b.item.Table.Schema == "" {
		b.item.Table.Schema = "public"
	}
	if !b.item.Table.RLSEnabled && !b.hasState {
		b.item.Table.RLSEnabled = true
	}
	if !b.item.Table.RLSForced && !b.hasState {
		b.item.Table.RLSForced = false
	}
}

func (b *tableBuilder) processFields() {
	for i := 0; i < b.modelType.NumField(); i++ {
		field := b.modelType.Field(i)
		if b.shouldSkip(field.Name) {
			continue
		}
		b.collectColumn(field)
	}
}

func (b *tableBuilder) collectColumn(field reflect.StructField) {
	columnTag := field.Tag.Get("column")
	if columnTag == "" {
		return
	}

	ct := raiden.UnmarshalColumnTag(columnTag)
	column := objects.Column{
		Table:  b.item.Table.Name,
		Schema: b.item.Table.Schema,
	}
	if existing, ok := b.existingColumns[ct.Name]; ok {
		column = existing
	}

	bindColumn(&field, &ct, &column)
	column.Table = b.item.Table.Name
	column.Schema = b.item.Table.Schema

	if column.DataType == string(postgres.TextType) {
		if _, exist := b.mapDataType[ct.Type]; exist {
			column.DataType = string(postgres.UserDefined)
			column.Format = ct.Type
		}
	}

	if ct.PrimaryKey {
		column.IsUnique = false
		b.appendPrimaryKey(column)
	}

	if vTag := field.Tag.Get("validate"); len(vTag) > 0 {
		b.item.ValidationTags[column.Name] = vTag
	}

	b.columns = append(b.columns, column)
}

func (b *tableBuilder) collectModelRelations() {
	for i := 0; i < b.modelType.NumField(); i++ {
		field := b.modelType.Field(i)
		if b.shouldSkip(field.Name) {
			continue
		}
		if join := field.Tag.Get("join"); join != "" {
			b.addModelRelation(field, join)
		}
	}
}

func (b *tableBuilder) collectStateRelations() {
	for i := 0; i < b.modelType.NumField(); i++ {
		field := b.modelType.Field(i)
		if b.shouldSkip(field.Name) {
			continue
		}
		if join := field.Tag.Get("join"); join != "" {
			b.addStateRelation(field, join)
		}
	}
}

func (b *tableBuilder) applyPolicies() {
	// TODO : remove after check and test
	// aclField, ok := b.modelType.FieldByName("Acl")
	// if !ok {
	// 	return
	// }

	// policies := getPolicies(&aclField, &b.item)
	// if len(policies) == 0 {
	// 	return
	// }

	if b.acl == nil {
		return
	}

	if b.item.Table.RLSEnabled != b.acl.IsEnable() {
		b.item.Table.RLSEnabled = b.acl.IsEnable()
	}

	if b.item.Table.RLSForced != b.acl.IsForced() {
		b.item.Table.RLSForced = b.acl.IsForced()
	}

	policies, err := b.acl.BuildPolicies(b.item.Table.Schema, b.item.Table.Name)
	if err != nil {
		panic(err.Error())
	}

	for _, p := range policies {
		if existing, ok := b.existingPolicies[p.Name]; ok {
			existing.Roles = p.Roles
			existing.Check = p.Check
			existing.Definition = p.Definition
			b.item.ExtractedPolicies.Existing = append(b.item.ExtractedPolicies.Existing, existing)
			delete(b.existingPolicies, p.Name)
			continue
		}
		b.item.ExtractedPolicies.New = append(b.item.ExtractedPolicies.New, p)
	}

}

func (b *tableBuilder) finish() ExtractTableItem {
	b.item.Table.Columns = b.columns
	b.item.Table.Relationships = b.relations
	b.item.Table.PrimaryKeys = b.primaryKeys

	if len(b.existingPolicies) > 0 {
		for _, p := range b.existingPolicies {
			b.item.ExtractedPolicies.Delete = append(b.item.ExtractedPolicies.Delete, p)
		}
	}

	return b.item
}

func (b *tableBuilder) addModelRelation(field reflect.StructField, join string) {
	jt := raiden.UnmarshalJoinTag(join)
	if jt.JoinType != raiden.RelationTypeHasOne {
		return
	}

	relation := objects.TablesRelationship{
		SourceTableName:   b.item.Table.Name,
		SourceSchema:      b.item.Table.Schema,
		TargetTableName:   utils.ToSnakeCase(field.Name),
		TargetTableSchema: b.item.Table.Schema,
	}

	if jt.ForeignKey != "" {
		relation.SourceColumnName = jt.ForeignKey
	} else {
		relation.SourceColumnName = fmt.Sprintf("%s_id", utils.ToSnakeCase(b.item.Table.Name))
	}

	if jt.PrimaryKey != "" {
		relation.TargetColumnName = jt.PrimaryKey
	} else {
		relation.TargetColumnName = "id"
	}

	b.applyRelationActions(&relation, field)
	if relation.Action != nil && relation.Action.ConstraintName == "" {
		relation.Action.ConstraintName = fmt.Sprintf("%s_%s_fkey", relation.SourceTableName, relation.SourceColumnName)
	}

	b.relations = append(b.relations, relation)
}

func (b *tableBuilder) addStateRelation(field reflect.StructField, join string) {
	tableName := findTypeName(field.Type, reflect.Struct, 4)
	if tableName == "" {
		return
	}

	relation := buildTableRelation(b.item.Table.Name, tableName, b.item.Table.Schema, b.existingRelations, join)
	if relation.ConstraintName == "" {
		return
	}

	b.applyRelationActions(&relation, field)
	b.relations = append(b.relations, relation)
}

func (b *tableBuilder) applyRelationActions(relation *objects.TablesRelationship, field reflect.StructField) {
	onUpdate := strings.ToLower(field.Tag.Get("onUpdate"))
	onDelete := strings.ToLower(field.Tag.Get("onDelete"))
	if onUpdate == "" && onDelete == "" {
		return
	}

	if relation.Action == nil {
		relation.Action = &objects.TablesRelationshipAction{}
	}

	if onUpdate == "" {
		relation.Action.UpdateAction = string(objects.RelationActionDefault)
	} else if v, ok := objects.RelationActionMapCode[objects.RelationActionLabel(onUpdate)]; ok {
		relation.Action.UpdateAction = string(v)
	} else {
		relation.Action.UpdateAction = string(objects.RelationActionDefault)
	}

	if onDelete == "" {
		relation.Action.DeletionAction = string(objects.RelationActionDefault)
	} else if v, ok := objects.RelationActionMapCode[objects.RelationActionLabel(onDelete)]; ok {
		relation.Action.DeletionAction = string(v)
	} else {
		relation.Action.DeletionAction = string(objects.RelationActionDefault)
	}

	if relation.Action.ConstraintName == "" {
		relation.Action.ConstraintName = fmt.Sprintf("%s_%s_fkey", relation.SourceTableName, relation.SourceColumnName)
	}

	if relation.Action.SourceSchema == "" {
		relation.Action.SourceSchema = relation.SourceSchema
	}
	if relation.Action.SourceTable == "" {
		relation.Action.SourceTable = relation.SourceTableName
	}
	if relation.Action.SourceColumns == "" {
		relation.Action.SourceColumns = fmt.Sprintf("{%s}", relation.SourceColumnName)
	}
	if relation.Action.TargetSchema == "" {
		relation.Action.TargetSchema = relation.TargetTableSchema
	}
	if relation.Action.TargetTable == "" {
		relation.Action.TargetTable = relation.SourceTableName
	}
	if relation.Action.TargetColumns == "" {
		relation.Action.TargetColumns = fmt.Sprintf("{%s}", relation.TargetColumnName)
	}
}

func (b *tableBuilder) appendPrimaryKey(column objects.Column) {
	if column.Name == "" {
		return
	}

	if existing, ok := b.existingPrimaryKey[column.Name]; ok {
		existing.Schema = column.Schema
		existing.TableName = b.item.Table.Name
		b.primaryKeys = append(b.primaryKeys, existing)
		return
	}

	pk := objects.PrimaryKey{
		Name:      column.Name,
		Schema:    column.Schema,
		TableName: b.item.Table.Name,
	}
	if b.item.Table.ID != 0 {
		pk.TableID = b.item.Table.ID
	}
	b.primaryKeys = append(b.primaryKeys, pk)
}

func (b *tableBuilder) shouldSkip(fieldName string) bool {
	return fieldName == "Metadata" || fieldName == "Acl"
}

func bindColumn(field *reflect.StructField, ct *raiden.ColumnTag, c *objects.Column) {
	c.IsNullable = ct.Nullable
	c.IsUnique = ct.Unique

	if ct.Name != "" {
		c.Name = ct.Name
	} else {
		c.Name = utils.ToSnakeCase(field.Name)
		ct.Name = c.Name
	}

	if ct.Type != "" {
		pgType := postgres.GetPgDataTypeName(postgres.DataType(ct.Type), false)
		c.DataType = string(pgType)
	} else {
		c.DataType = string(postgres.ToPostgresType(field.Type.Name()))
	}

	c.DefaultValue = ct.Default

	if ct.AutoIncrement {
		c.IsIdentity = true
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
	} else {
		if r, ok := mapRelations[getRelationConstrainNameWithoutSchema(sourceTableName, foreignKey)]; ok {
			relation = r
		}
	}

	relation.SourceSchema = schema
	relation.SourceTableName = sourceTableName
	relation.SourceColumnName = foreignKey

	relation.TargetTableSchema = schema
	relation.TargetTableName = targetTableName
	relation.TargetColumnName = primaryKey

	return
}

// get Acl
func getAcl(model any) *raiden.Acl {
	v := reflect.ValueOf(model)
	if !v.IsValid() {
		return nil
	}

	// ensure pointer to struct
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Kind() != reflect.Pointer {
		if v.CanAddr() {
			v = v.Addr()
		} else {
			p := reflect.New(v.Type())
			p.Elem().Set(v)
			v = p
		}
	}
	if v.IsNil() {
		return nil
	}
	elem := v.Elem()
	if elem.Kind() != reflect.Struct {
		return nil
	}

	f := elem.FieldByName("Acl")
	if f.IsValid() {
		// model has: Acl raiden.Acl   (value)  or  Acl *raiden.Acl   (pointer)
		switch f.Kind() {
		case reflect.Struct:
			if f.CanAddr() {
				if a, ok := f.Addr().Interface().(*raiden.Acl); ok {
					callConfigureOnce(v, a)
					return a
				}
			}
		case reflect.Pointer:
			if f.IsNil() {
				// allocate if nil so ConfigureAcl can mutate it
				p := reflect.New(f.Type().Elem())
				if initializer, ok := p.Interface().(*raiden.Acl); ok && initializer != nil {
					initializer.Define() // ensure internal map gets initialised via guard
				}
				f.Set(p)
			}
			current := f.Interface()
			if a, ok := current.(*raiden.Acl); ok && a != nil {
				callConfigureOnce(v, a)
				// Configuration might have swapped the pointer; read the field again
				if updated, ok := f.Interface().(*raiden.Acl); ok && updated != nil {
					return updated
				}
				return a
			}
		}
	}

	return nil
}

func callConfigureOnce(modelVal reflect.Value, a *raiden.Acl) {
	if a == nil {
		return
	}
	// If model has ConfigureAcl(), run it exactly once for this ACL instance
	if m := modelVal.MethodByName("ConfigureAcl"); m.IsValid() && m.Type().NumIn() == 0 && m.Type().NumOut() == 0 {
		a.InitOnce(func() { m.Call(nil) })
	}
}

// get relation table name, base on struct type that defined in relation field
func getRelationConstrainName(schema, table, foreignKey string) string {
	return fmt.Sprintf("%s_%s_%s_fkey", schema, table, foreignKey)
}

func getRelationConstrainNameWithoutSchema(table, foreignKey string) string {
	return fmt.Sprintf("%s_%s_fkey", table, foreignKey)
}

func (f ExtractTableItems) ToFlatTable() (tables []objects.Table) {
	for i := range f {
		t := f[i]
		tables = append(tables, t.Table)
	}
	return
}

func (f ExtractTableResult) ToDeleteFlatMap() map[string]*objects.Table {
	mapData := make(map[string]*objects.Table)

	if len(f.Delete) > 0 {
		for i := range f.Delete {
			r := f.Delete[i]
			mapData[r.Table.Name] = &r.Table
		}
	}

	return mapData
}

func findTypeName(sf reflect.Type, findType reflect.Kind, maxDeep int) string {
	if maxDeep == 0 {
		return ""
	}

	if sf.Kind() == findType {
		return sf.Name()
	}

	switch sf.Kind() {
	case reflect.Ptr:
		return findTypeName(sf.Elem(), findType, maxDeep-1)
	case reflect.Array:
		return findTypeName(sf.Elem(), findType, maxDeep-1)
	default:
		return ""
	}
}
