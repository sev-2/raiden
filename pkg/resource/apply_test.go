package resource_test

import (
	"testing"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/sev-2/raiden/pkg/resource"
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/resource/policies"
	"github.com/sev-2/raiden/pkg/resource/roles"
	"github.com/sev-2/raiden/pkg/resource/rpc"
	"github.com/sev-2/raiden/pkg/resource/storages"
	"github.com/sev-2/raiden/pkg/resource/tables"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func loadConfig() *raiden.Config {
	return &raiden.Config{
		DeploymentTarget:    raiden.DeploymentTargetCloud,
		ProjectId:           "test-project-id",
		ProjectName:         "test-project",
		SupabaseApiBasePath: "/v1",
		SupabaseApiUrl:      "http://supabase.cloud.com",
		SupabasePublicUrl:   "http://supabase.cloud.com",
	}
}

func TestApply(t *testing.T) {
	flags := &resource.Flags{
		DryRun:        true,
		AllowedSchema: "public",
	}
	config := loadConfig()

	err := resource.Apply(flags, config)
	assert.Error(t, err)

	flags.DryRun = false
	importState := &state.LocalState{
		State: state.State{
			Tables: []state.TableState{
				{
					Table: objects.Table{
						Name:        "test_local_table",
						PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
						Columns: []objects.Column{
							{Name: "id", DataType: "uuid"},
							{Name: "name", DataType: "text"},
						},
						Relationships: []objects.TablesRelationship{
							{
								ConstraintName:    "test_local_constraint",
								SourceSchema:      "public",
								SourceTableName:   "test_local_table",
								SourceColumnName:  "id",
								TargetTableSchema: "public",
								TargetTableName:   "test_table",
								TargetColumnName:  "id",
							},
						},
					},
				},
			},
			Storage: []state.StorageState{
				{
					Storage: objects.Bucket{
						Name:   "test_bucket_policy",
						Public: true,
					},
				},
			},
		},
	}

	errSaveState := state.Save(&importState.State)
	assert.NoError(t, errSaveState)

	mock := &mock.MockSupabase{Cfg: config}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetRolesWithExpectedResponse(200, []objects.Role{})
	assert.NoError(t, err0)

	err1 := mock.MockGetBucketsWithExpectedResponse(200, []objects.Bucket{})
	assert.NoError(t, err1)

	err = resource.Apply(flags, config)
	assert.NoError(t, err)

	errReset := state.Save(&state.State{})
	assert.NoError(t, errReset)
}

func TestMigrate(t *testing.T) {
	config := loadConfig()
	importState := &state.LocalState{
		State: state.State{
			Tables: []state.TableState{
				{
					Table: objects.Table{
						Name:        "test_local_table",
						PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
						Columns: []objects.Column{
							{Name: "id", DataType: "uuid"},
							{Name: "name", DataType: "text"},
						},
						Relationships: []objects.TablesRelationship{
							{
								ConstraintName:    "test_local_constraint",
								SourceSchema:      "public",
								SourceTableName:   "test_local_table",
								SourceColumnName:  "id",
								TargetTableSchema: "public",
								TargetTableName:   "test_table",
								TargetColumnName:  "id",
							},
						},
					},
				},
			},
			Storage: []state.StorageState{
				{
					Storage: objects.Bucket{
						Name:   "test_bucket_policy",
						Public: true,
					},
				},
			},
		},
	}

	errSaveState := state.Save(&importState.State)
	assert.NoError(t, errSaveState)

	projectPath := "/path/to/project"
	resources := &resource.MigrateData{
		Tables: []tables.MigrateItem{
			{
				Type: migrator.MigrateTypeCreate,
				NewData: objects.Table{
					Name:        "test_table",
					PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
					Columns: []objects.Column{
						{Name: "id", DataType: "uuid"},
						{Name: "name", DataType: "text"},
					},
					Relationships: []objects.TablesRelationship{
						{
							ConstraintName:    "test_constraint",
							SourceSchema:      "public",
							SourceTableName:   "test_table",
							SourceColumnName:  "id",
							TargetTableSchema: "public",
							TargetTableName:   "test_table",
							TargetColumnName:  "id",
						},
					},
				},
			},
			{
				Type: migrator.MigrateTypeUpdate,
				OldData: objects.Table{
					Name:        "test_table",
					PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
					Columns: []objects.Column{
						{Name: "id", DataType: "uuid"},
						{Name: "name", DataType: "text"},
					},
				},
				NewData: objects.Table{
					Name:        "test_table",
					PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
					Columns: []objects.Column{
						{Name: "id", DataType: "uuid"},
						{Name: "name", DataType: "text"},
						{Name: "age", DataType: "integer"},
					},
					Relationships: []objects.TablesRelationship{
						{
							ConstraintName:    "test_constraint",
							SourceSchema:      "public",
							SourceTableName:   "test_table",
							SourceColumnName:  "id",
							TargetTableSchema: "public",
							TargetTableName:   "test_table",
							TargetColumnName:  "id",
						},
					},
				},
			},
			{
				Type: migrator.MigrateTypeDelete,
				OldData: objects.Table{
					Name:        "test_table_deleted",
					PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
					Columns: []objects.Column{
						{Name: "id", DataType: "uuid"},
						{Name: "name", DataType: "text"},
					},
				},
			},
		},
		Roles: []roles.MigrateItem{
			{
				Type: migrator.MigrateTypeCreate,
				NewData: objects.Role{
					Name: "test_role",
				},
			},
			{
				Type: migrator.MigrateTypeUpdate,
				OldData: objects.Role{
					Name:     "test_role",
					CanLogin: false,
				},
				NewData: objects.Role{
					Name:     "test_role",
					CanLogin: true,
				},
			},
			{
				Type: migrator.MigrateTypeDelete,
				OldData: objects.Role{
					Name: "test_role_deleted",
				},
			},
		},
		Rpc: []rpc.MigrateItem{
			{
				Type: migrator.MigrateTypeCreate,
				NewData: objects.Function{
					Name:              "test_rpc",
					CompleteStatement: "test_rpc()",
				},
			},
			{
				Type: migrator.MigrateTypeUpdate,
				OldData: objects.Function{
					Name:              "test_rpc",
					CompleteStatement: "test_rpc()",
				},
				NewData: objects.Function{
					Name:              "test_rpc",
					CompleteStatement: "test_rpc_updated()",
				},
			},
			{
				Type: migrator.MigrateTypeDelete,
				OldData: objects.Function{
					Name:              "test_rpc_deleted",
					CompleteStatement: "test_rpc_deleted()",
				},
			},
		},
		Policies: []policies.MigrateItem{
			{
				Type: migrator.MigrateTypeCreate,
				NewData: objects.Policy{
					Name:   "test_policy",
					Schema: "public",
					Table:  "test_table",
				},
			},
			{
				Type: migrator.MigrateTypeCreate,
				OldData: objects.Policy{
					Name:   "test_policy_created",
					Schema: "public",
					Table:  "test_table",
				},
				NewData: objects.Policy{
					Name:   "storage test_bucket_policy",
					Schema: supabase.DefaultStorageSchema,
					Table:  supabase.DefaultObjectTable,
				},
			},
			{
				Type: migrator.MigrateTypeUpdate,
				OldData: objects.Policy{
					Name:   "test_policy1",
					Schema: "public",
					Table:  "test_table",
				},
				NewData: objects.Policy{
					Name:   "test_policy1",
					Schema: "public",
					Table:  "test_table",
					Action: "SELECT",
				},
			},
			{
				Type: migrator.MigrateTypeUpdate,
				OldData: objects.Policy{
					Name:   "test_policy2",
					Schema: "public",
					Table:  "test_table",
				},
				NewData: objects.Policy{
					Name:   "storage test_bucket_policy",
					Schema: supabase.DefaultStorageSchema,
					Table:  supabase.DefaultObjectTable,
					Action: "SELECT",
				},
			},
			{
				Type: migrator.MigrateTypeDelete,
				OldData: objects.Policy{
					Name:   "test_deleted_policy",
					Schema: "public",
					Table:  "test_table_deleted",
				},
			},
			{
				Type: migrator.MigrateTypeDelete,
				OldData: objects.Policy{
					Name:   "storage test_bucket_policy",
					Schema: supabase.DefaultStorageSchema,
					Table:  supabase.DefaultObjectTable,
				},
			},
		},
		Storages: []storages.MigrateItem{
			{
				Type: migrator.MigrateTypeCreate,
				NewData: objects.Bucket{
					Name:   "test_bucket",
					Public: true,
				},
			},
			{
				Type: migrator.MigrateTypeUpdate,
				OldData: objects.Bucket{
					Name:   "test_bucket1",
					Public: true,
				},
				NewData: objects.Bucket{
					Name:   "test_bucket1",
					Public: false,
				},
			},
			{
				Type: migrator.MigrateTypeDelete,
				OldData: objects.Bucket{
					Name:   "test_bucket_deleted",
					Public: true,
				},
			},
		},
	}

	mock := &mock.MockSupabase{Cfg: config}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetTablesWithExpectedResponse(200, []objects.Table{{Name: "test_table"}})
	assert.NoError(t, err0)

	err1 := mock.MockCreateBucketsWithExpectedResponse(200, objects.Bucket{Name: "test_bucket"})
	assert.NoError(t, err1)

	err2 := mock.MockDeleteBucketsWithExpectedResponse(200)
	assert.NoError(t, err2)

	err3 := mock.MockGetBucketByNameWithExpectedResponse(200, objects.Bucket{Name: "test_bucket"})
	assert.NoError(t, err3)

	errs := resource.Migrate(config, importState, projectPath, resources)
	assert.Empty(t, errs)

	errReset := state.Save(&state.State{})
	assert.NoError(t, errReset)
}

func TestUpdateLocalStateFromApply(t *testing.T) {
	projectPath := "/path/to/project"
	localState := &state.LocalState{}
	stateChan := make(chan any)
	done := resource.UpdateLocalStateFromApply(projectPath, localState, stateChan)

	go func() {
		defer close(stateChan)
		stateChan <- &tables.MigrateItem{
			Type:    migrator.MigrateTypeCreate,
			NewData: objects.Table{Name: "test_table"},
		}
	}()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for UpdateLocalStateFromApply to complete")
	}
}

func TestPrintApplyChangeReport(t *testing.T) {
	migrateData := resource.MigrateData{
		Tables: []tables.MigrateItem{
			{Type: migrator.MigrateTypeCreate, NewData: objects.Table{Name: "test_table"}},
		},
	}

	resource.PrintApplyChangeReport(migrateData)
}
