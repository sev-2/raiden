package imports

import (
	"fmt"
	"sync"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
)

// The `generateResource` function generates various resources such as table, roles, policy and etc
// also generate framework resource like controller, route, main function and etc
func generateResource(config *raiden.Config, importState *ImportState, projectPath string, resource *Resource) error {
	if err := generator.CreateInternalFolder(projectPath); err != nil {
		return err
	}

	wg, errChan, stateChan := sync.WaitGroup{}, make(chan error), make(chan any)
	doneListen := ListenStateResource(importState, stateChan)

	// generate all model from cloud / pg-meta
	if len(resource.Tables) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			inputs := buildGenerateModelInputs(resource.Tables)
			captureFunc := StateDecorateFunc(inputs, func(item *generator.GenerateModelInput, input generator.GenerateInput) bool {
				if i, ok := input.BindData.(generator.GenerateModelData); ok {
					if i.StructName == utils.SnakeCaseToPascalCase(item.Table.Name) {
						return true
					}
				}
				return false
			}, stateChan)

			if err := generator.GenerateModels(projectPath, inputs, resource.Policies, captureFunc); err != nil {
				errChan <- err
			} else {
				errChan <- nil
			}
		}()
	}

	// generate all roles from cloud / pg-meta
	if len(resource.Roles) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			captureFunc := StateDecorateFunc(resource.Roles, func(item supabase.Role, input generator.GenerateInput) bool {
				if i, ok := input.BindData.(generator.GenerateRoleData); ok {
					if i.RoleName == item.Name {
						return true
					}
				}
				return false
			}, stateChan)

			if err := generator.GenerateRoles(projectPath, resource.Roles, captureFunc); err != nil {
				errChan <- err
			} else {
				errChan <- nil
			}
		}()
	}

	// TODO : generate rpc
	if len(resource.Functions) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// TODO : generate all function from cloud / pg-meta
			errChan <- nil
		}()
	}

	go func() {
		wg.Wait()
		close(stateChan)
		close(errChan)
	}()

	for {
		select {
		case rsErr := <-errChan:
			if rsErr != nil {
				return rsErr
			}
		case saveErr := <-doneListen:
			return saveErr
		}
	}
}

func buildGenerateModelInputs(tables []supabase.Table) []*generator.GenerateModelInput {
	mapTable := tableToMap(tables)
	mapRelations := buildMapRelations(mapTable)
	return buildGenerateModelInput(mapTable, mapRelations)
}

// ----- Map Relations -----

func scanTableRelation(table *supabase.Table) (relations []*generator.Relation, manyToManyCandidates []*ManyToManyTable) {
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

		relation := generator.Relation{
			Table:        tableName,
			Type:         typePrefix + utils.SnakeCaseToPascalCase(tableName),
			RelationType: relationType,
			PrimaryKey:   primaryKey,
			ForeignKey:   foreignKey,
		}

		relations = append(relations, &relation)
	}

	return
}

func mergeRelations(table *supabase.Table, relations []*generator.Relation, mapRelations MapRelations) {
	key := getMapTableKey(table.Schema, table.Name)
	tableRelations, isExist := mapRelations[key]
	if isExist {
		tableRelations = append(tableRelations, relations...)
	} else {
		tableRelations = relations
	}
	mapRelations[key] = tableRelations
}

func mergeManyToManyCandidate(candidates []*ManyToManyTable, mapRelations MapRelations) {
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
				rs = make([]*generator.Relation, 0)
			}

			r := generator.Relation{
				Table:        targetTable.Table,
				Type:         "[]*" + utils.SnakeCaseToPascalCase(targetTable.Table),
				RelationType: raiden.RelationTypeManyToMany,
				JoinRelation: &generator.JoinRelation{
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

func buildMapRelations(mapTable MapTable) MapRelations {
	mr := make(MapRelations)
	for _, t := range mapTable {
		r, m2m := scanTableRelation(t)
		if len(r) == 0 {
			continue
		}

		// merge with existing relation
		mergeRelations(t, r, mr)

		// merge many to many candidate with table relations
		mergeManyToManyCandidate(m2m, mr)
	}
	return mr
}

// --- attach relation to table
func buildGenerateModelInput(mapTable MapTable, mapRelations MapRelations) []*generator.GenerateModelInput {
	generateInputs := make([]*generator.GenerateModelInput, 0)
	for k, v := range mapTable {
		input := generator.GenerateModelInput{
			Table: *v,
		}

		if r, exist := mapRelations[k]; exist {
			for _, v := range r {
				if v != nil {
					v.Tag = v.BuildTag()
					input.Relations = append(input.Relations, *v)
				}
			}
		}

		generateInputs = append(generateInputs, &input)
	}
	return generateInputs
}

func tableToMap(tables []supabase.Table) MapTable {
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
