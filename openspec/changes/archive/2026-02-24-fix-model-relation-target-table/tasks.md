## 1. Implementation
- [x] 1.1 Fix `addModelRelation` to use `raiden.GetTableName()` on the field's type for `TargetTableName` instead of `field.Name`

## 2. Testing
- [x] 2.1 Run state tests (`go test ./pkg/state/`)
- [x] 2.2 Run resource tests (`go test ./pkg/resource/...`)
- [x] 2.3 Validate in pivot project with `apply --dry-run`
