## MODIFIED Requirements

### Requirement: Relation Comparison Accuracy
The relation comparison SHALL correctly match source and target relations even when the
database contains duplicate FK constraints for the same column or cross-schema FK references.

#### Scenario: Duplicate FK constraints on same column
- **WHEN** the remote database has two FK constraints for the same source column (e.g., custom-named `fk_mc_division` and default-named `master_creators_division_id_fkey`)
- **AND** the local code has one FK for that column
- **THEN** the matched constraint SHALL be recognized as identical
- **AND** the duplicate constraint SHALL NOT be flagged as a delete

#### Scenario: Cross-schema FK reference
- **WHEN** the remote database has a FK referencing a table in a different schema (e.g., `public.user_brands.user_id → auth.users.id`)
- **AND** the local code does not represent this FK (because the target table is not in the imported model set)
- **THEN** the cross-schema FK SHALL NOT be flagged as a delete

#### Scenario: Index creation check
- **WHEN** both local and remote sides have no index for a relation
- **THEN** no index creation SHALL be proposed
- **WHEN** the remote has an index but the local does not
- **THEN** an index creation item SHALL be proposed

### Requirement: Policy Comparison Accuracy
The policy comparison SHALL match policies by their full identity (schema, table, and name)
rather than by name alone, to prevent cross-table mismatches when multiple tables share
the same policy name.

#### Scenario: Same-named policies on different tables
- **WHEN** multiple tables have policies with the same name (e.g., "admin full access" on `products` and `product_interaction_performance`)
- **THEN** each policy SHALL be compared only with its corresponding policy on the same table
- **AND** no false table-change diffs SHALL be reported

### Requirement: RPC Comparison Accuracy
The RPC state extraction SHALL preserve the CompleteStatement from the previous import
rather than rebuilding it from the Go struct template, to avoid format-only differences
between `BuildRpc()` output and `pg_get_functiondef()` output.

#### Scenario: No code changes after import
- **WHEN** a user runs import and then apply without modifying any RPC code
- **THEN** no RPC updates SHALL be detected
