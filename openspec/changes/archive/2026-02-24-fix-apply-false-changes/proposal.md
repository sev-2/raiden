# Change: Fix apply dry-run false-change detection

## Why
After running `import --force` followed by `apply --dry-run`, the system incorrectly
reports relation creates/deletes, policy table reassignments, and RPC updates even
though no code was modified. This makes the dry-run output noisy and untrustworthy.

## What Changes
- Fix relation comparison to skip duplicate FK constraints for already-matched columns
- Fix relation comparison to skip cross-schema FK references (e.g., `auth.users`)
- Fix relation index creation check to only fire when target has index but source doesn't
- Fix policy comparison to match by schema+table+name instead of name only
- Fix RPC state extraction to preserve stored CompleteStatement from previous import

## Impact
- Affected specs: cli-imports
- Affected code:
  - `pkg/resource/tables/compare.go` (relation comparison)
  - `pkg/resource/policies/compare.go` (policy comparison)
  - `pkg/state/rpc.go` (RPC state extraction)
