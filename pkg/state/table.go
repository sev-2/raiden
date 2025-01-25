package state

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type ModelValidationTag map[string]string

type ExtractTableItem struct {
	Table             objects.Table
	ValidationTags    ModelValidationTag
	ExtractedPolicies ExtractedPolicies
}

type ExtractTableItems []ExtractTableItem

type ExtractTableResult struct {
	Existing ExtractTableItems
	New      ExtractTableItems
	Delete   ExtractTableItems
}

func ExtractTable(tableStates []TableState, appTable []any) (result ExtractTableResult, err error) {
	var mapTableState = make(map[string]TableState)

	for i := range tableStates {
		t := tableStates[i]
		mapTableState[t.Table.Name] = t
	}

	for _, t := range appTable {
		tableName := raiden.GetTableName(t)
		ts, isExist := mapTableState[tableName]

		if !isExist {
			nt := buildTableFromModel(t)
			result.New = append(result.New, nt)
			continue
		}

		tb := buildTableFromState(t, ts)
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
			ExtractedPolicies: ExtractedPolicies{
				Delete: deletedPolicy,
			},
		})
	}

	return
}

func buildTableFromModel(model any) (ei ExtractTableItem) {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	ei.Table.Name = raiden.GetTableName(model)
	ei.ValidationTags = make(ModelValidationTag)

	// add metadata
	metadataField, isExist := modelType.FieldByName("Metadata")
	if isExist {
		bindTableMetadata(&metadataField, &ei.Table)
	} else {
		ei.Table.Schema = "public"
		ei.Table.RLSEnabled = true
		ei.Table.RLSForced = false
	}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		switch field.Name {
		case "Metadata", "Acl":
			continue
		default:
			if column := field.Tag.Get("column"); len(column) > 0 {
				ct := raiden.UnmarshalColumnTag(column)

				c := objects.Column{
					Table:    ei.Table.Name,
					Schema:   ei.Table.Schema,
					IsUnique: false,
				}

				bindColumn(&field, &ct, &c)
				ei.Table.Columns = append(ei.Table.Columns, c)

				if ct.PrimaryKey {
					ei.Table.PrimaryKeys = append(ei.Table.PrimaryKeys, objects.PrimaryKey{
						Name:      c.Name,
						TableName: c.Table,
						Schema:    c.Schema,
					})

					// set is unique to false if is primary key
					c.IsUnique = false
				}

				if vTag := field.Tag.Get("validate"); len(vTag) > 0 {
					ei.ValidationTags[c.Name] = vTag
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

					// check action field
					onUpdate := field.Tag.Get("onUpdate")
					onDelete := field.Tag.Get("onDelete")
					if len(onUpdate) > 0 || len(onDelete) > 0 {
						if len(onUpdate) == 0 {
							onUpdate = string(objects.RelationActionDefault)
						} else {
							v, ok := objects.RelationActionMapCode[objects.RelationActionLabel(onUpdate)]
							if ok {
								onUpdate = string(v)
							}
						}

						if len(onDelete) == 0 {
							onDelete = string(objects.RelationActionDefault)
						} else {
							v, ok := objects.RelationActionMapCode[objects.RelationActionLabel(onDelete)]
							if ok {
								onDelete = string(v)
							}
						}

						rel.Action = &objects.TablesRelationshipAction{
							ConstraintName: fmt.Sprintf("%s_%s_fkey", rel.SourceTableName, rel.SourceColumnName),
							UpdateAction:   onUpdate,
							DeletionAction: onDelete,
							SourceSchema:   rel.SourceSchema,
							SourceTable:    rel.SourceTableName,
							SourceColumns:  fmt.Sprintf("{%s}", rel.SourceColumnName),

							TargetSchema:  rel.TargetTableSchema,
							TargetTable:   rel.SourceTableName,
							TargetColumns: fmt.Sprintf("{%s}", rel.TargetColumnName),
						}
					}

					// check index field
					indexed := field.Tag.Get("indexed")
					if len(indexed) > 0 {
						rel.Index = &objects.Index{
							Schema:     ei.Table.Schema,
							Table:      rel.TargetTableName,
							Name:       fmt.Sprintf("ix_%s_%s", rel.TargetTableName, rel.TargetColumnName),
							Definition: fmt.Sprintf("CREATE INDEX ix_%s_%s ON %s(%s);", rel.TargetTableName, rel.TargetColumnName, rel.TargetTableName, rel.TargetColumnName),
						}
					}

					ei.Table.Relationships = append(ei.Table.Relationships, rel)
				}
			}
		}
	}

	// add metadata
	aclField, isExist := modelType.FieldByName("Acl")
	if isExist {
		ei.ExtractedPolicies.New = getPolicies(&aclField, &ei)
	}

	return
}

func buildTableFromState(model any, state TableState) (ei ExtractTableItem) {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// Get the reflect.Type of the struct
	ei.Table = state.Table
	ei.Table.Name = raiden.GetTableName(model)
	ei.ValidationTags = make(ModelValidationTag)

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

	mapPolicies := make(map[string]objects.Policy)
	for ip := range state.Policies {
		p := state.Policies[ip]
		mapPolicies[p.Name] = p
	}

	var columns []objects.Column
	var relations []objects.TablesRelationship
	var primaryKeys []objects.PrimaryKey

	// update metadata
	metadataField, isExist := modelType.FieldByName("Metadata")
	if isExist {
		bindTableMetadata(&metadataField, &ei.Table)
	} else {
		ei.Table.Schema = "public"
		ei.Table.RLSEnabled = true
		ei.Table.RLSForced = false
	}

	// Iterate over the fields of the struct
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		switch field.Name {
		case "Metadata", "Acl":
			continue
		default:
			// example tag "name:id;type:bigint;primaryKey;autoIncrement;nullable:false"
			if columnTag := field.Tag.Get("column"); len(columnTag) > 0 {
				var c objects.Column
				c.IsUnique = false

				ct := raiden.UnmarshalColumnTag(columnTag)
				if found, exist := mapColumn[ct.Name]; exist {
					c = found
				}

				c.Table = ei.Table.Name
				c.Schema = ei.Table.Schema

				bindColumn(&field, &ct, &c)

				if ct.PrimaryKey {
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

				if vTag := field.Tag.Get("validate"); len(vTag) > 0 {
					ei.ValidationTags[c.Name] = vTag
				}

				columns = append(columns, c)
			}

			if joinTag := field.Tag.Get("join"); len(joinTag) > 0 {
				tableName := findTypeName(field.Type, reflect.Struct, 4)
				if tableName != "" {
					if r := buildTableRelation(ei.Table.Name, tableName, ei.Table.Schema, mapRelation, joinTag); r.ConstraintName != "" {
						if onUpdate := field.Tag.Get("onUpdate"); onUpdate != "" {
							if r.Action == nil {
								r.Action = &objects.TablesRelationshipAction{}
							}

							r.Action.UpdateAction = string(objects.RelationActionMapCode[objects.RelationActionLabel(strings.ToLower(onUpdate))])
						}

						if onDelete := field.Tag.Get("onDelete"); onDelete != "" {
							if r.Action == nil {
								r.Action = &objects.TablesRelationshipAction{}
							}

							r.Action.DeletionAction = string(objects.RelationActionMapCode[objects.RelationActionLabel(strings.ToLower(onDelete))])
						}

						relations = append(relations, r)
					}
				}

			}
		}
	}

	// add acl
	aclField, isExist := modelType.FieldByName("Acl")
	if isExist {
		policies := getPolicies(&aclField, &ei)
		for ip := range policies {
			p := policies[ip]

			sp, exist := mapPolicies[p.Name]
			if !exist {
				ei.ExtractedPolicies.New = append(ei.ExtractedPolicies.New, p)
				continue
			}

			sp.Roles = p.Roles
			sp.Check = p.Check
			sp.Definition = p.Definition
			ei.ExtractedPolicies.Existing = append(ei.ExtractedPolicies.Existing, sp)
			delete(mapPolicies, p.Name)
		}
	}

	if len(mapPolicies) > 0 {
		for _, v := range mapPolicies {
			ei.ExtractedPolicies.Delete = append(ei.ExtractedPolicies.Delete, v)
		}
	}

	ei.Table.Columns = columns
	ei.Table.Relationships = relations
	ei.Table.PrimaryKeys = primaryKeys

	return ei
}

func bindColumn(field *reflect.StructField, ct *raiden.ColumnTag, c *objects.Column) {
	c.IsNullable = ct.Nullable
	c.IsUnique = ct.Unique

	if ct.Name != "" {
		c.Name = ct.Name
	} else {
		ct.Name = utils.ToSnakeCase(field.Name)
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

func getPolicies(field *reflect.StructField, ei *ExtractTableItem) (policies []objects.Policy) {
	acl := raiden.UnmarshalAclTag(string(field.Tag))
	tableType := strings.ToLower(string(supabase.RlsTypeModel))

	defaultCheck, defaultDefinition := "true", "true"
	if len(acl.Read.Roles) > 0 {
		readPolicyName := supabase.GetPolicyName(objects.PolicyCommandSelect, tableType, ei.Table.Name)
		policy := objects.Policy{
			Name:       readPolicyName,
			Schema:     ei.Table.Schema,
			Table:      ei.Table.Name,
			Action:     "PERMISSIVE",
			Command:    objects.PolicyCommandSelect,
			Roles:      acl.Read.Roles,
			Definition: acl.Read.Using,
		}
		if policy.Definition == "" {
			policy.Definition = defaultDefinition
		} else if policy.Definition != defaultDefinition {
			policy.Definition = fmt.Sprintf("(%s)", policy.Definition)
		}

		policies = append(policies, policy)
	}

	if len(acl.Write.Roles) > 0 {
		createPolicy := objects.Policy{
			Name:    supabase.GetPolicyName(objects.PolicyCommandInsert, tableType, ei.Table.Name),
			Schema:  ei.Table.Schema,
			Table:   ei.Table.Name,
			Action:  "PERMISSIVE",
			Command: objects.PolicyCommandInsert,
			Roles:   acl.Write.Roles,
			Check:   acl.Write.Check,
		}
		if createPolicy.Check == nil || (createPolicy.Check != nil && *createPolicy.Check == "") {
			createPolicy.Check = &defaultCheck
		} else if createPolicy.Check != nil && *createPolicy.Check != "" && *createPolicy.Check != defaultCheck {
			check := fmt.Sprintf("(%s)", *createPolicy.Check)
			createPolicy.Check = &check
		}

		updatePolicy := objects.Policy{
			Name:       supabase.GetPolicyName(objects.PolicyCommandUpdate, tableType, ei.Table.Name),
			Schema:     ei.Table.Schema,
			Table:      ei.Table.Name,
			Action:     "PERMISSIVE",
			Command:    objects.PolicyCommandUpdate,
			Roles:      acl.Write.Roles,
			Definition: acl.Write.Using,
			Check:      acl.Write.Check,
		}
		if updatePolicy.Check == nil || (updatePolicy.Check != nil && *updatePolicy.Check == "") {
			updatePolicy.Check = &defaultCheck
		} else if updatePolicy.Check != nil && *updatePolicy.Check != "" && *updatePolicy.Check != defaultCheck {
			check := fmt.Sprintf("(%s)", *updatePolicy.Check)
			updatePolicy.Check = &check
		}

		if updatePolicy.Definition == "" {
			updatePolicy.Definition = "true"
		} else if updatePolicy.Definition != defaultDefinition {
			updatePolicy.Definition = fmt.Sprintf("(%s)", updatePolicy.Definition)
		}

		deletePolicy := objects.Policy{
			Name:       supabase.GetPolicyName(objects.PolicyCommandDelete, tableType, ei.Table.Name),
			Schema:     ei.Table.Schema,
			Table:      ei.Table.Name,
			Action:     "PERMISSIVE",
			Command:    objects.PolicyCommandDelete,
			Roles:      acl.Write.Roles,
			Definition: acl.Write.Using,
		}
		if deletePolicy.Definition == "" {
			deletePolicy.Definition = "true"
		} else if deletePolicy.Definition != defaultDefinition {
			deletePolicy.Definition = fmt.Sprintf("(%s)", deletePolicy.Definition)
		}
		policies = append(policies, createPolicy, updatePolicy, deletePolicy)
	}

	return
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
