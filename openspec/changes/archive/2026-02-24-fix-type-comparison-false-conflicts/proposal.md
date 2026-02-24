# Change: Fix false-positive conflicts during import comparison

## Why
The import comparison logic produces false-positive conflicts that block `raiden imports` after a clean `--force` import across all three resource comparison subsystems:

1. **Type comparison**: Compares `*string` pointer addresses instead of values for `Comment` fields; iterates slice indices instead of values for `Enums` and `Attributes`.
2. **Table relation comparison**: Constraint name mismatch between remote (real DB names like `fk_mc_division`) and local (generated names like `public_table_col_fkey`) causes relations to be flagged as new. Missing remote `Action` data treated as conflict during import. Cross-schema FK references (e.g., `auth.users`) not filtered before comparison.
3. **RPC function comparison**: `BindRpcFunction` overwrites state's stored `CompleteStatement` with rebuilt one from `BuildRpc()`, which differs from `pg_get_functiondef()` in parameter prefix (`in_`), default value quoting (`'null'::uuid` vs `null::uuid`), and `search_path` inclusion.

## What Changes
### Type comparison (`pkg/resource/types/compare.go`)
- Fix `Comment` pointer comparison to dereference and compare string values
- Remove duplicate `Comment` comparison block
- Fix `Enums` and `Attributes` range loops to iterate values instead of indices

### Table relation comparison (`pkg/resource/tables/compare.go`)
- Add secondary index `mapTargetByCol` keyed by `schema.table.column` for fallback relation matching when constraint name lookup fails
- Guard relation `Action` nil check to only flag as conflict in apply mode, not import mode

### Table state extraction (`pkg/state/table.go`)
- Add fallback relation lookup by `SourceTableName+SourceColumnName` in `buildTableRelation` when constraint name lookup fails

### Cross-schema relation filtering (`pkg/resource/import.go`)
- Filter remote relationships to exclude references to tables not in the local model set before comparison, mirroring generator behavior that skips cross-schema FKs

### RPC comparison (`pkg/resource/import.go`)
- Restore state's stored `CompleteStatement` before comparison instead of using rebuilt one from `BuildRpc()`, so import only detects real remote changes

## Impact
- Affected specs: cli-imports
- Affected code: `pkg/resource/types/compare.go`, `pkg/resource/tables/compare.go`, `pkg/state/table.go`, `pkg/resource/import.go`
