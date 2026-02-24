# CLI Imports

## Purpose

The `raiden imports` CLI command synchronises a local Raiden project with a remote Supabase database by fetching resources (tables, roles, RPC functions, storage buckets, custom types, and policies), comparing them against the local state, and generating Go source files that represent those resources as code. It uses a two-stage build approach: a temporary import binary is code-generated with the user's registered bootstrap code, compiled, and executed to perform the actual import.

**Key files:** `cmd/raiden/commands/import.go`, `pkg/cli/imports/command.go`, `pkg/resource/import.go`, `pkg/resource/load.go`, `pkg/resource/common.go`, `pkg/generator/import.go`, `pkg/state/state.go`
## Requirements
### Requirement: CLI Command Registration

The system SHALL register an `imports` cobra subcommand on the root CLI that accepts resource selection flags, behaviour flags, and debug flags. The command SHALL run a `PreRun` check to verify that `configs/app.yaml` exists before proceeding.

#### Scenario: Command invocation with valid config
- **WHEN** the user runs `raiden imports`
- **THEN** the system SHALL load the project configuration from `configs/app.yaml` and proceed with the import workflow

#### Scenario: Command invocation without config file
- **WHEN** the user runs `raiden imports` and `configs/app.yaml` does not exist
- **THEN** the system SHALL return an error: "missing config file (./configs/app.yaml), run `raiden configure` first for generate configuration file"

#### Scenario: Version check on invocation
- **WHEN** the user runs `raiden imports`
- **THEN** the system SHALL check for CLI updates before proceeding, and if an update is available and applied, SHALL exit with code 0

### Requirement: Resource Selection Flags

The system SHALL support flags to selectively import specific resource types. When no resource-specific flag is provided, all resource types SHALL be imported.

#### Scenario: Import all resources (default)
- **WHEN** the user runs `raiden imports` without `--models-only`, `--rpc-only`, `--roles-only`, `--storages-only`, or `--policy-only`
- **THEN** the system SHALL import tables, roles, RPC functions, storage buckets, types, and policies

#### Scenario: Import models only
- **WHEN** the user runs `raiden imports --models-only`
- **THEN** the system SHALL import only tables/models and their associated types

#### Scenario: Import RPC only
- **WHEN** the user runs `raiden imports --rpc-only`
- **THEN** the system SHALL import only RPC functions (and load types and tables for parameter resolution)

#### Scenario: Import roles only
- **WHEN** the user runs `raiden imports --roles-only`
- **THEN** the system SHALL import only user-defined roles

#### Scenario: Import storages only
- **WHEN** the user runs `raiden imports --storages-only`
- **THEN** the system SHALL import only storage buckets

#### Scenario: Import policies only
- **WHEN** the user runs `raiden imports --policy-only`
- **THEN** the system SHALL import only policies

### Requirement: Schema Filtering

The system SHALL filter tables and functions by the schemas specified via `--schema`. When no schema is specified, the system SHALL default to `public`.

#### Scenario: Default schema filter
- **WHEN** the user runs `raiden imports` without `--schema`
- **THEN** only tables and functions in the `public` schema SHALL be included

#### Scenario: Multiple schema filter
- **WHEN** the user runs `raiden imports --schema auth,public,storage`
- **THEN** tables and functions in the `auth`, `public`, and `storage` schemas SHALL be included, and all others SHALL be excluded

#### Scenario: Relation validation during filtering
- **WHEN** a table has relationships referencing tables outside the imported schema set
- **THEN** the system SHALL log a debug warning identifying the missing relation target but SHALL NOT fail the import

### Requirement: Allowed Tables Filter (BFF Mode)

When running in BFF mode with `config.AllowedTables` set to a value other than `"*"`, the system SHALL further restrict imported tables to only those listed in `AllowedTables`. Relationships referencing tables not in the allowed list SHALL be removed.

#### Scenario: BFF mode with restricted tables
- **WHEN** the config mode is `bff` and `AllowedTables` is `"users,orders"`
- **THEN** only the `users` and `orders` tables SHALL be imported, and relationships to other tables SHALL be stripped

#### Scenario: BFF mode with wildcard
- **WHEN** the config mode is `bff` and `AllowedTables` is `"*"`
- **THEN** all tables matching the schema filter SHALL be imported without further restriction

### Requirement: Two-Stage Build Execution

The system SHALL use a two-stage approach: first code-generate a `cmd/import/main.go` binary that embeds the user's registered models, then compile and execute that binary as a subprocess to perform the actual import.

#### Scenario: Binary compilation
- **WHEN** `imports.Run()` is called
- **THEN** the system SHALL execute `go build -o build/import cmd/import/main.go`, deleting any previously built binary first

#### Scenario: Binary execution with flags
- **WHEN** the import binary is compiled successfully
- **THEN** the system SHALL execute `build/import` with the appropriate flags forwarded (e.g., `--models-only`, `--schema`, `--force`, `--dry-run`, `--debug`)

#### Scenario: Windows platform support
- **WHEN** the target OS is `windows`
- **THEN** the output binary path SHALL have a `.exe` extension

#### Scenario: Build failure
- **WHEN** `go build` fails (e.g., syntax error in generated code)
- **THEN** the system SHALL return an error: "error building binary: {details}"

### Requirement: Generated Import Binary Bootstrap

The code-generated `cmd/import/main.go` SHALL register the user's application resources (models, types, and optionally RPC, roles, storages in BFF mode) via `bootstrap.Register*()` calls, then run a pre-generate pass, call `resource.Import()`, and run a post-generate pass to refresh bootstrap files.

#### Scenario: BFF mode bootstrap
- **WHEN** the config mode is `bff`
- **THEN** the generated binary SHALL call `bootstrap.RegisterModels()`, `bootstrap.RegisterTypes()`, `bootstrap.RegisterRpc()`, `bootstrap.RegisterRoles()`, and `bootstrap.RegisterStorages()`

#### Scenario: Service mode bootstrap
- **WHEN** the config mode is not `bff`
- **THEN** the generated binary SHALL call only `bootstrap.RegisterModels()` and `bootstrap.RegisterTypes()`

#### Scenario: Post-import regeneration
- **WHEN** the import completes successfully and `--dry-run` is not set
- **THEN** the generated binary SHALL run `generate.Run()` a second time to regenerate bootstrap files reflecting newly imported resources

### Requirement: Concurrent Remote Resource Loading

The `Load()` function SHALL fetch all required resources from the remote Supabase database concurrently using goroutines. Results SHALL be collected through a typed channel.

#### Scenario: Concurrent fetch in BFF mode
- **WHEN** `Load()` is called in BFF mode with all resources
- **THEN** tables, roles, role memberships, functions, storages, indexes, relation actions, policies, and types SHALL be fetched concurrently via `pkg/supabase` API calls

#### Scenario: Concurrent fetch in Service mode
- **WHEN** `Load()` is called in Service mode
- **THEN** tables, functions, indexes, relation actions, and types SHALL be fetched from PgMeta (`pkg/connector/pgmeta`), and roles from the Supabase API

#### Scenario: Fetch error propagation
- **WHEN** any resource fetch goroutine encounters an error
- **THEN** the error SHALL be sent through the channel and `Load()` SHALL return that error immediately

### Requirement: Post-Load Resource Enrichment

After loading remote resources, the system SHALL attach additional metadata to the raw resource data.

#### Scenario: Table enrichment
- **WHEN** tables are loaded
- **THEN** `tables.AttachIndexAndAction()` SHALL attach indexes and relationship actions to their respective tables

#### Scenario: Role enrichment
- **WHEN** roles are loaded
- **THEN** `roles.AttachInherithRole()` SHALL attach inherited role memberships to each role, using the native role map to resolve references

### Requirement: Native Role Handling

The system SHALL maintain a map of built-in PostgreSQL/Supabase roles (e.g., `postgres`, `supabase_admin`, `pg_*`) and use it to separate native roles from user-defined roles.

#### Scenario: Native role exclusion from import
- **WHEN** remote roles are loaded
- **THEN** native roles SHALL be excluded from the importable roles list (only user-defined roles are imported as code)

#### Scenario: Native role state tracking
- **WHEN** remote roles include native roles
- **THEN** native roles SHALL be recorded in the import state for reference but SHALL NOT be code-generated

### Requirement: Local State Management

The system SHALL persist import state in the `.raiden/` directory using Go's GOB binary encoding. The state tracks previously imported tables, roles, RPC functions, storage buckets, and types.

#### Scenario: First import (no existing state)
- **WHEN** no `.raiden/` state exists
- **THEN** all remote resources SHALL be treated as new and code-generated

#### Scenario: Subsequent import
- **WHEN** `.raiden/` state exists from a prior import
- **THEN** the system SHALL load the state, extract registered resources, and classify each as New or Existing

#### Scenario: State update during generation
- **WHEN** a resource file is generated
- **THEN** the resource metadata and output path SHALL be sent through `stateChan` to `UpdateLocalStateFromImport()`, which updates the `LocalState` in real-time

#### Scenario: State persistence
- **WHEN** all code generation is complete (the `stateChan` channel closes)
- **THEN** `LocalState.Persist()` SHALL write the updated state to `.raiden/`

### Requirement: Resource Extraction and Classification

The system SHALL compare the local state against currently registered Go resources (from `bootstrap.Register*()`) and classify each resource as either New (in state but not registered) or Existing (in both state and code).

#### Scenario: Extract tables
- **WHEN** `--models-only` or all resources are being imported
- **THEN** `state.ExtractTable()` SHALL classify tables into New and Existing sets

#### Scenario: Extract roles
- **WHEN** `--roles-only` or all resources are being imported
- **THEN** `state.ExtractRole()` SHALL classify roles into New and Existing sets

#### Scenario: Extract RPC functions
- **WHEN** `--rpc-only` or all resources are being imported
- **THEN** `state.ExtractRpc()` SHALL classify RPC functions into New and Existing sets

#### Scenario: Extract storages
- **WHEN** `--storages-only` or all resources are being imported
- **THEN** `state.ExtractStorage()` SHALL classify storage buckets into New and Existing sets

### Requirement: Validation Tag Preservation

When importing models, the system SHALL preserve existing model validation tags (e.g., `validate:"required"`) from both New and Existing table extractions so that regeneration does not lose user-defined validation constraints.

#### Scenario: Preserve validation tags during reimport
- **WHEN** an existing model has validation tags defined
- **THEN** the regenerated model file SHALL include those same validation tags

### Requirement: Comparison and Diff Checks

The system SHALL compare remote (Supabase) resources against existing local resources to detect drift, unless `--force` is set. Pointer-typed fields (e.g., `*string`) SHALL be compared by dereferenced value, not by pointer address. Slice fields SHALL be compared by element values, not by slice indices. Relation action comparisons SHALL only flag a conflict when action data is available from both sides, or when running in apply mode. Relation matching SHALL fall back to `schema.table.column` lookup when constraint name lookup fails. Cross-schema FK references SHALL be filtered out before comparison when the target table is not in the local model set. RPC `CompleteStatement` comparison during import SHALL use the stored state value, not the rebuilt value from `BuildRpc()`.

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

### Requirement: Import Report

The system SHALL compute and display a report showing the count of new resources added per type (Table, Role, Rpc, Storage, Types, Policies).

#### Scenario: Report after successful import
- **WHEN** the import generates new resources
- **THEN** the system SHALL log: "import process is complete, adding several new resources to the codebase" with counts for each resource type

#### Scenario: Report when no changes
- **WHEN** the import finds no new resources
- **THEN** the system SHALL log: "import process is complete, your code is up to date"

#### Scenario: Dry-run report with no errors
- **WHEN** `--dry-run` completes without comparison errors
- **THEN** the system SHALL log: "finish running import in dry run mode" with resource counts

#### Scenario: Dry-run report with errors
- **WHEN** `--dry-run` completes with comparison errors
- **THEN** the system SHALL log the collected errors and skip the resource count report

#### Scenario: Report printed exactly once
- **WHEN** the import workflow completes
- **THEN** the report SHALL be printed exactly once (guarded by `reportPrinted` flag)

### Requirement: Output Handling Modes

The system SHALL support three output modes based on flags: normal (generate code), dry-run (report only), and update-state-only (update `.raiden/` without generating code).

#### Scenario: Normal output (code generation)
- **WHEN** neither `--dry-run` nor `--update-state-only` is set
- **THEN** the system SHALL run `generateImportResource()` to create Go source files and update state

#### Scenario: Dry-run output
- **WHEN** `--dry-run` is set
- **THEN** the system SHALL NOT write any files; it SHALL only print the report or collected errors

#### Scenario: Update state only
- **WHEN** `--update-state-only` is set
- **THEN** the system SHALL update `.raiden/` state with the remote resource data without generating Go source files

### Requirement: Concurrent Code Generation

The `generateImportResource()` function SHALL generate all resource types concurrently within a single goroutine group, sending state updates through a channel as each file is generated.

#### Scenario: Generate types
- **WHEN** custom PostgreSQL types exist in the remote resource
- **THEN** `generator.GenerateTypes()` SHALL produce type files under `internal/types/`

#### Scenario: Generate models
- **WHEN** tables exist in the remote resource
- **THEN** `generator.GenerateModels()` SHALL produce model structs under `internal/models/` with column tags, join tags, and preserved validation tags

#### Scenario: Generate REST controllers (optional)
- **WHEN** `--generate-controller` is set and config mode is `bff`
- **THEN** `generator.GenerateRestControllers()` SHALL produce controller stubs for each imported table

#### Scenario: Generate roles
- **WHEN** user-defined roles exist in the remote resource
- **THEN** `generator.GenerateRoles()` SHALL produce role definitions under `internal/roles/`

#### Scenario: Generate RPC functions
- **WHEN** functions exist in the remote resource
- **THEN** `generator.GenerateRpc()` SHALL produce RPC wrappers under `internal/rpc/`

#### Scenario: Generate storages
- **WHEN** storage buckets exist in the remote resource
- **THEN** `generator.GenerateStorages()` SHALL produce storage definitions under `internal/storages/`

#### Scenario: Generation error handling
- **WHEN** any generator function returns an error
- **THEN** the error SHALL be sent through `errChan` and the import SHALL return that error

### Requirement: Import Binary Code Generation

The system SHALL generate a `cmd/import/main.go` file from a Go template that embeds the user's project module name and bootstrap imports.

#### Scenario: Template rendering
- **WHEN** `GenerateImportMainFunction()` is called
- **THEN** it SHALL create `cmd/import/main.go` with the correct package imports, bootstrap registration calls, and cobra command setup

#### Scenario: Directory creation
- **WHEN** the `cmd/` or `cmd/import/` directories do not exist
- **THEN** the system SHALL create them before writing the file

#### Scenario: Import paths
- **WHEN** the template is rendered
- **THEN** it SHALL include imports for `raiden`, `generate`, `imports`, `resource`, `utils`, `cobra`, and the user's `internal/bootstrap` package

### Requirement: Dependency Injection for Testability

The `importJob` struct SHALL use an `importDeps` struct that holds function references for all external dependencies (loading, comparing, generating, reporting). This allows tests to stub individual phases.

#### Scenario: Default dependencies
- **WHEN** `Import()` is called in production
- **THEN** `defaultImportDeps` SHALL be used, wiring real implementations for all phases

#### Scenario: Test dependencies
- **WHEN** `runImport()` is called in tests
- **THEN** custom `importDeps` MAY be provided to stub any phase (e.g., `loadRemote`, `compareTables`)

### Requirement: Panic Recovery

The `runImport()` function SHALL include a deferred panic recovery that converts any panic into a returned error.

#### Scenario: Panic during import
- **WHEN** any phase panics
- **THEN** the system SHALL recover and return an error: "import panic: {details}"

#### Scenario: Deferred report on success
- **WHEN** the import completes without error and the report has been computed but not yet printed
- **THEN** the deferred function SHALL print the report before returning

### Requirement: Pre-Import Generation Pass

Before the core import logic runs, the system SHALL execute `generate.Run()` to refresh internal bootstrap artifacts. This ensures the import binary has up-to-date type registrations.

#### Scenario: Bootstrap refresh before import
- **WHEN** the CLI command runs the import workflow
- **THEN** `generate.Run()` SHALL be called before `imports.Run()` to regenerate route files and bootstrap registrations

#### Scenario: Generation failure before import
- **WHEN** `generate.Run()` fails before the import
- **THEN** the system SHALL log the error and abort without proceeding to the import

