# Change: Fix model relation TargetTableName derivation

## Why
`addModelRelation` in `pkg/state/table.go` derives `TargetTableName` from the struct **field name** (`utils.ToSnakeCase(field.Name)`) instead of the referenced **type name**. When a field name differs from the type name (e.g., field `MasterCreatorBrand` of type `*MasterCreators`), the relation points to a non-existent table (`master_creator_brand` instead of `master_creators`). This causes `validateTableRelations` during `apply` to fail with "target column id is not exist in table master_creator_brand".

## What Changes
- Fix `addModelRelation` in `pkg/state/table.go` to resolve `TargetTableName` from the field's type name using `findTypeName()`, matching the behavior of `addStateRelation`

## Impact
- Affected specs: cli-imports
- Affected code: `pkg/state/table.go`
