# Imports Command – Structural Diagram

This class-style view highlights the key packages/files involved in `raiden imports` and where external Supabase/PgMeta calls originate.

```mermaid
classDiagram
    class RaidenCLI {
        <<package:cmd/raiden>>
        +main.go
        +cobra.Command rootCmd
    }

    class ImportCommand {
        <<file:cmd/raiden/commands/import.go>>
        +Run(cmd,*cobra.Command,args []string)
        +ImportsFlags
    }

    class GenerateRunner {
        <<pkg:pkg/cli/generate>>
        +Run(flags,*raiden.Config,string,bool) error
    }

    class ImportsRunner {
        <<pkg:pkg/cli/imports>>
        +Run(logFlags,*Flags,string) error
        +PreRun(projectPath string) error
    }

    class ImportBinary {
        <<file:cmd/import/main.go>>
        +main()
        +resource.Flags
    }

    class ResourceImporter {
        <<pkg:pkg/resource/import.go>>
        +Import(*Flags,*raiden.Config) error
    }

    class ResourceLoader {
        <<pkg:pkg/resource/load.go>>
        +Load(*Flags,*raiden.Config) (*Resource,error)
    }

    class StateStore {
        <<pkg:pkg/state>>
        +Load() (*LocalState,error)
    }

    class SupabaseClient {
        <<pkg:pkg/supabase>>
        +GetTables(cfg,*raiden.Config,schemas[]string)
        +GetFunctions(cfg)
        +GetRoles(cfg)
        +GetBuckets(cfg)
    }

    class PgMetaConnector {
        <<pkg:pkg/connector/pgmeta>>
        +GetTables(cfg)
        +GetIndexes(cfg)
        +GetTableRelationshipActions(cfg)
    }

    class SupabaseDrivers {
        <<pkg:pkg/supabase/drivers>>
        +cloud.GetTables()
        +cloud.GetFunctions()
        +local/meta.GetTables()
    }

    RaidenCLI --> ImportCommand : registers
    ImportCommand --> GenerateRunner : bootstrap refresh
    ImportCommand --> ImportsRunner : kicks off import build/exec
    ImportsRunner --> ImportBinary : exec build/import
    ImportBinary --> GenerateRunner : pre/post regeneration
    ImportBinary --> ResourceImporter : resource.Import()
    ResourceImporter --> ResourceLoader : Load()
    ResourceImporter --> StateStore : Load()
    ResourceLoader --> SupabaseClient : BFF mode
    ResourceLoader --> PgMetaConnector : Service mode
    SupabaseClient --> SupabaseDrivers : driver selection
```

- `pkg/supabase` delegates HTTP calls to `pkg/supabase/drivers/cloud` (Supabase SQL/PostgREST) or `pkg/supabase/drivers/local/meta`.
- `pkg/connector/pgmeta` issues HTTP requests to PgMeta when running in service mode, feeding schema metadata back to the loader.
