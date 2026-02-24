## MODIFIED Requirements

### Requirement: Comparison and Diff Checks

The system SHALL compare remote (Supabase) resources against existing local resources to detect drift, unless `--force` is set. Pointer-typed fields (e.g., `*string`) SHALL be compared by dereferenced value, not by pointer address. Slice fields SHALL be compared by element values, not by slice indices. Relation action comparisons SHALL only flag a conflict when action data is available from both sides, or when running in apply mode. Relation matching SHALL fall back to `schema.table.column` lookup when constraint name lookup fails. Cross-schema FK references SHALL be filtered out before comparison when the target table is not in the local model set. RPC `CompleteStatement` comparison during import SHALL use the stored state value, not the rebuilt value from `BuildRpc()`. Model relation `TargetTableName` SHALL be derived from the referenced type name, not from the struct field name, to ensure it matches the actual database table.

#### Scenario: Normal comparison
- **WHEN** `--force` is not set and existing resources are present
- **THEN** the system SHALL run comparison checks for types, tables, roles, RPC functions, and storages

#### Scenario: Comparison error in normal mode
- **WHEN** a comparison detects conflicting changes
- **THEN** the system SHALL return an error and abort the import

#### Scenario: Comparison error in dry-run mode
- **WHEN** `--dry-run` is set and a comparison detects conflicts
- **THEN** the system SHALL collect the error message without aborting, and report it at the end

#### Scenario: Skip comparisons with force flag
- **WHEN** `--force` is set
- **THEN** all comparison checks SHALL be skipped and remote state SHALL overwrite local unconditionally

#### Scenario: Skip comparison when no existing resources
- **WHEN** the Existing set for a resource type is empty (first import or new resource type)
- **THEN** the comparison for that resource type SHALL be skipped

#### Scenario: No false conflict on identical pointer-typed fields
- **WHEN** a type's `Comment` field has the same string value on both remote and local but different pointer addresses
- **THEN** the comparison SHALL report no conflict

#### Scenario: No false conflict on identical enum values
- **WHEN** a type's `Enums` slice has the same string values on both remote and local
- **THEN** the comparison SHALL report no conflict regardless of slice allocation

#### Scenario: No false conflict on identical attribute values
- **WHEN** a type's `Attributes` slice has the same `Name` and `TypeID` values on both remote and local
- **THEN** the comparison SHALL report no conflict regardless of slice allocation

#### Scenario: No false conflict on missing remote relation action during import
- **WHEN** a relation's remote `Action` is nil (not attached) but local `Action` is populated, and the comparison mode is import
- **THEN** the comparison SHALL NOT flag this as a conflict

#### Scenario: Flag missing action as conflict in apply mode
- **WHEN** a relation's local `Action` is populated but remote `Action` is nil, and the comparison mode is apply
- **THEN** the comparison SHALL flag this as a conflict for `OnUpdate` and `OnDelete` actions

#### Scenario: No false conflict on constraint name mismatch
- **WHEN** a relation exists in both remote and local with different constraint names but identical `schema.table.column` reference
- **THEN** the comparison SHALL match them via fallback lookup and report no conflict

#### Scenario: No false conflict on cross-schema FK references
- **WHEN** a remote table has a FK referencing a table in a different schema (e.g., `auth.users`) that is not in the local model set
- **THEN** the comparison SHALL exclude that relationship before comparison

#### Scenario: No false conflict on RPC CompleteStatement formatting
- **WHEN** an RPC function's `CompleteStatement` from `pg_get_functiondef()` differs from the `BuildRpc()` rebuilt version only in formatting (param prefix, default quoting, search_path)
- **THEN** the import comparison SHALL use the stored state `CompleteStatement` and report no conflict

#### Scenario: Correct TargetTableName for model relations
- **WHEN** a model struct field references a type with a different name than the field (e.g., field `MasterCreatorBrand` of type `*MasterCreators`)
- **THEN** the relation `TargetTableName` SHALL be derived from the type name (`master_creators`), not the field name (`master_creator_brand`)
