## 1. Type comparison fixes
- [x] 1.1 Fix `Comment` pointer comparison in `CompareItem()` to dereference `*string` values
- [x] 1.2 Remove duplicate `Comment` comparison block (lines 106-115)
- [x] 1.3 Fix `Enums` range loop to iterate values (`for _, se := range`) instead of indices
- [x] 1.4 Fix `Attributes` range loop to iterate values and compare `Name`/`TypeID` fields

## 2. Table relation comparison fixes
- [x] 2.1 Add `mapTargetByCol` secondary index in `compareRelations()` for fallback matching by `schema.table.column`
- [x] 2.2 Guard relation `Action` nil check to only flag diff in apply mode (not import)
- [x] 2.3 Add fallback relation lookup in `buildTableRelation()` by `SourceTableName+SourceColumnName`
- [x] 2.4 Add cross-schema relation filtering in `compareTables()` to exclude FKs referencing tables not in local model set

## 3. RPC comparison fix
- [x] 3.1 Restore state `CompleteStatement` in `importJob.compareRpc()` before comparison instead of using rebuilt value from `BuildRpc()`

## 4. Testing
- [x] 4.1 Run existing type comparison tests (`go test ./pkg/resource/types/`)
- [x] 4.2 Run table comparison tests (`go test ./pkg/resource/tables/`)
- [x] 4.3 Run state tests (`go test ./pkg/state/`)
- [x] 4.4 Run import tests (`go test ./pkg/resource/`)
- [x] 4.5 Run RPC tests (`go test ./pkg/resource/rpc/`)
- [x] 4.6 Validate in pivot project — zero false-positive conflicts
