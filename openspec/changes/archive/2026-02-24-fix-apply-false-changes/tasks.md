## 1. Investigation
- [x] 1.1 Trace relation false creates/deletes — root cause: duplicate FK constraints + cross-schema FKs
- [x] 1.2 Trace policy false table reassignment — root cause: CompareList maps by name only, not schema+table+name
- [x] 1.3 Trace RPC false updates — root cause: BindRpcFunction overwrites state CompleteStatement with BuildRpc() output

## 2. Implementation
- [x] 2.1 Add `matchedSourceCols` tracking to skip duplicate FK deletes in `compareRelations`
- [x] 2.2 Add cross-schema FK filter to skip FKs where TargetTableSchema != SourceSchema
- [x] 2.3 Change index creation condition from `t.Index == nil && sc.Index == nil` to `t.Index != nil && sc.Index == nil`
- [x] 2.4 Fix policy `CompareList` to use schema+table+name key instead of name only
- [x] 2.5 Preserve state CompleteStatement in `ExtractRpc` for existing functions

## 3. Validation
- [x] 3.1 All existing tests pass
- [x] 3.2 Pivot project: `import` shows zero conflicts
- [x] 3.3 Pivot project: `apply --dry-run` shows only genuinely new resources
