package tables

import (
	"fmt"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

// ----- Convert array of table to map table -----
type MapTable map[string]*objects.Table

func tableToMap(tables []objects.Table) MapTable {
	mapTable := make(MapTable)
	for i := range tables {
		t := tables[i]
		key := getMapTableKey(t.Schema, t.Name)
		mapTable[key] = &t
	}
	return mapTable
}

func getMapTableKey(schema, name string) string {
	return fmt.Sprintf("%s.%s", schema, name)
}

func BuildGenerateModelInputs(tables []objects.Table, policies objects.Policies, mapModelValidationTags map[string]state.ModelValidationTag) []*generator.GenerateModelInput {
	mapTable := tableToMap(tables)
	mapRelations := buildGenerateMapRelations(mapTable)
	return buildGenerateModelInput(mapTable, mapRelations, policies, mapModelValidationTags)
}

// ---- build table relation for generated -----
type (
	MapRelations    map[string][]*state.Relation
	ManyToManyTable struct {
		Table      string
		Schema     string
		PivotTable string
		PrimaryKey string
		ForeignKey string
	}
)

func buildGenerateMapRelations(mapTable MapTable) MapRelations {
	mr := make(MapRelations)
	for _, t := range mapTable {
		r, m2m := scanGenerateTableRelation(t)
		if len(r) == 0 {
			continue
		}

		// merge with existing relation
		mergeGenerateRelations(t, r, mr)

		// merge many to many candidate with table relations
		mergeGenerateManyToManyCandidate(m2m, mr)
	}
	return mr
}

func scanGenerateTableRelation(table *objects.Table) (relations []*state.Relation, manyToManyCandidates []*ManyToManyTable) {
	// skip process if doesn`t have relation`
	if len(table.Relationships) == 0 {
		return
	}

	for _, r := range table.Relationships {
		var tableName string
		var primaryKey = r.TargetColumnName
		var foreignKey = r.SourceColumnName
		var typePrefix = "*"
		var relationType = raiden.RelationTypeHasMany

		if r.SourceTableName == table.Name {
			relationType = raiden.RelationTypeHasOne
			tableName = r.TargetTableName

			// hasOne relation is candidate to many to many relation
			// assumption table :
			//  table :
			// 		- teacher
			// 		- topic
			// 		- class
			// 	relation :
			// 		- teacher has many class
			// 		- topic has many class
			// 		- class has one teacher and has one topic
			manyToManyCandidates = append(manyToManyCandidates, &ManyToManyTable{
				Table:      r.TargetTableName,
				PivotTable: table.Name,
				PrimaryKey: r.TargetColumnName,
				ForeignKey: r.SourceColumnName,
				Schema:     r.TargetTableSchema,
			})
		} else {
			typePrefix = "[]*"
			tableName = r.SourceTableName
		}

		relation := state.Relation{
			Table:        tableName,
			Type:         typePrefix + utils.SnakeCaseToPascalCase(tableName),
			RelationType: relationType,
			PrimaryKey:   primaryKey,
			ForeignKey:   foreignKey,
			Action:       r.Action,
			Index:        r.Index,
		}

		relations = append(relations, &relation)
	}

	return
}

func mergeGenerateRelations(table *objects.Table, relations []*state.Relation, mapRelations MapRelations) {
	key := getMapTableKey(table.Schema, table.Name)
	tableRelations, isExist := mapRelations[key]
	if isExist {
		tableRelations = append(tableRelations, relations...)
	} else {
		tableRelations = relations
	}
	mapRelations[key] = tableRelations
}

func mergeGenerateManyToManyCandidate(candidates []*ManyToManyTable, mapRelations MapRelations) {
	for sourceTableIndex, sourceTable := range candidates {
		for targetTableIndex, targetTable := range candidates {
			if sourceTableIndex == targetTableIndex {
				continue
			}

			if sourceTable == nil || targetTable == nil {
				continue
			}

			key := getMapTableKey(sourceTable.Schema, sourceTable.Table)
			rs, exist := mapRelations[key]
			if !exist {
				rs = make([]*state.Relation, 0)
			}

			r := state.Relation{
				Table:        targetTable.Table,
				Type:         "[]*" + utils.SnakeCaseToPascalCase(targetTable.Table),
				RelationType: raiden.RelationTypeManyToMany,
				JoinRelation: &state.JoinRelation{
					Through: sourceTable.PivotTable,

					SourcePrimaryKey:      sourceTable.PrimaryKey,
					JoinsSourceForeignKey: sourceTable.ForeignKey,

					TargetPrimaryKey:     targetTable.PrimaryKey,
					JoinTargetForeignKey: targetTable.ForeignKey,
				},
			}

			rs = append(rs, &r)
			mapRelations[key] = rs
		}

	}
}

// --- attach index and action to relation
func AttachIndexAndAction(allTable []objects.Table, allIndex []objects.Index, allAction []objects.TablesRelationshipAction) []objects.Table {
	// build map index
	mapIndex := make(map[string]objects.Index)
	for _, v := range allIndex {
		mapIndex[v.Name] = v
	}

	// build map action
	mapAction := make(map[string]objects.TablesRelationshipAction)
	for _, v := range allAction {
		key := fmt.Sprintf("%s_%s", v.SourceSchema, v.ConstraintName)
		mapAction[key] = v
	}

	for iTable := range allTable {
		table := allTable[iTable]
		for i := range table.Relationships {
			r := table.Relationships[i]

			// check index
			indexKey := fmt.Sprintf("ix_%s_%s", r.SourceTableName, r.SourceColumnName)
			if fIndex, exist := mapIndex[indexKey]; exist && fIndex.Name != "" {
				r.Index = &fIndex
			} else {
				indexKey := fmt.Sprintf("ix_%s_%s_%s", r.SourceSchema, r.SourceTableName, r.SourceColumnName)
				if fIndex2, exist := mapIndex[indexKey]; exist && fIndex2.Name != "" {
					r.Index = &fIndex2
				}
			}

			// check action
			if action, exist := mapAction[r.ConstraintName]; exist {
				r.Action = &action
			} else {
				actionKey := fmt.Sprintf("%s_%s", r.SourceSchema, r.ConstraintName)
				if action2, exist := mapAction[actionKey]; exist && action2.ConstraintName != "" {
					r.Action = &action2
				}
			}

			// replace with new value
			table.Relationships[i] = r
		}

		allTable[iTable] = table
	}

	return allTable
}

// --- attach relation to table
func buildGenerateModelInput(mapTable MapTable, mapRelations MapRelations, policies objects.Policies, mapModelValidationTags map[string]state.ModelValidationTag) []*generator.GenerateModelInput {
	generateInputs := make([]*generator.GenerateModelInput, 0)
	for k, v := range mapTable {
		input := generator.GenerateModelInput{
			Table:    *v,
			Policies: policies.FilterByTable(v.Name),
		}

		if r, exist := mapRelations[k]; exist {
			for _, v := range r {
				if v != nil {
					input.Relations = append(input.Relations, *v)
				}
			}
		}

		vTag, exist := mapModelValidationTags[input.Table.Name]
		if exist && vTag != nil {
			input.ValidationTags = vTag
		}
		generateInputs = append(generateInputs, &input)
	}
	return generateInputs
}
