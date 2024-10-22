package tables_test

import (
	"encoding/json"
	"testing"

	"github.com/sev-2/raiden/pkg/resource/tables"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestBuildGenerateModelInputs(t *testing.T) {
	jsonStrData := `[{"id":29072,"schema":"public","name":"candidate","rls_enabled":true,"rls_forced":false,"replica_identity":"DEFAULT","bytes":32768,"size":"32 kB","live_rows_estimate":2,"dead_rows_estimate":0,"comment":"list of candidate","columns":[{"table_id":29072,"schema":"public","table":"candidate","id":"29072.1","ordinal_position":1,"name":"id","default_value":null,"data_type":"bigint","format":"int8","is_identity":true,"identity_generation":"BY DEFAULT","is_generated":false,"is_nullable":false,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29072,"schema":"public","table":"candidate","id":"29072.2","ordinal_position":2,"name":"name","default_value":null,"data_type":"character varying","format":"varchar","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29072,"schema":"public","table":"candidate","id":"29072.3","ordinal_position":3,"name":"batch","default_value":null,"data_type":"bigint","format":"int8","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29072,"schema":"public","table":"candidate","id":"29072.4","ordinal_position":4,"name":"created_at","default_value":"now()","data_type":"timestamp with time zone","format":"timestamptz","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null}],"primary_keys":[{"schema":"public","table_name":"candidate","name":"id","table_id":29072}],"relationships":[{"id":29242,"constraint_name":"submission_candidate_id_fkey","source_schema":"public","source_table_name":"submission","source_column_name":"candidate_id","target_table_schema":"public","target_table_name":"candidate","target_column_name":"id"}]},{"id":29079,"schema":"public","name":"scouter","rls_enabled":true,"rls_forced":false,"replica_identity":"DEFAULT","bytes":16384,"size":"16 kB","live_rows_estimate":0,"dead_rows_estimate":0,"comment":"scouter list","columns":[{"table_id":29079,"schema":"public","table":"scouter","id":"29079.1","ordinal_position":1,"name":"id","default_value":null,"data_type":"bigint","format":"int8","is_identity":true,"identity_generation":"BY DEFAULT","is_generated":false,"is_nullable":false,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29079,"schema":"public","table":"scouter","id":"29079.2","ordinal_position":2,"name":"name","default_value":null,"data_type":"character varying","format":"varchar","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29079,"schema":"public","table":"scouter","id":"29079.3","ordinal_position":3,"name":"email","default_value":null,"data_type":"character varying","format":"varchar","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29079,"schema":"public","table":"scouter","id":"29079.4","ordinal_position":4,"name":"created_at","default_value":"now()","data_type":"timestamp with time zone","format":"timestamptz","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null}],"primary_keys":[{"schema":"public","table_name":"scouter","name":"id","table_id":29079}],"relationships":[{"id":30078,"constraint_name":"submission_scouter_id_fkey","source_schema":"public","source_table_name":"submission","source_column_name":"scouter_id","target_table_schema":"public","target_table_name":"scouter","target_column_name":"id"}]},{"id":29086,"schema":"public","name":"submission","rls_enabled":true,"rls_forced":false,"replica_identity":"DEFAULT","bytes":16384,"size":"16 kB","live_rows_estimate":0,"dead_rows_estimate":0,"comment":null,"columns":[{"table_id":29086,"schema":"public","table":"submission","id":"29086.1","ordinal_position":1,"name":"id","default_value":null,"data_type":"bigint","format":"int8","is_identity":true,"identity_generation":"BY DEFAULT","is_generated":false,"is_nullable":false,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29086,"schema":"public","table":"submission","id":"29086.2","ordinal_position":2,"name":"scouter_id","default_value":null,"data_type":"bigint","format":"int8","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29086,"schema":"public","table":"submission","id":"29086.3","ordinal_position":3,"name":"candidate_id","default_value":null,"data_type":"bigint","format":"int8","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29086,"schema":"public","table":"submission","id":"29086.4","ordinal_position":4,"name":"score","default_value":null,"data_type":"real","format":"float4","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29086,"schema":"public","table":"submission","id":"29086.5","ordinal_position":5,"name":"note","default_value":null,"data_type":"text","format":"text","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29086,"schema":"public","table":"submission","id":"29086.6","ordinal_position":6,"name":"created_at","default_value":"now()","data_type":"timestamp with time zone","format":"timestamptz","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null}],"primary_keys":[{"schema":"public","table_name":"submission","name":"id","table_id":29086}],"relationships":[{"id":29242,"constraint_name":"submission_candidate_id_fkey","source_schema":"public","source_table_name":"submission","source_column_name":"candidate_id","target_table_schema":"public","target_table_name":"candidate","target_column_name":"id","action":{"id":29242,"constraint_name":"submission_candidate_id_fkey","deletion_action":"NO ACTION","update_action":"NO ACTION","source_id":29086,"source_schema":"public","source_table":"submission","source_columns":"candidate_id","target_id":29072,"target_schema":"public","target_table":"candidate","target_columns":"id"},"index":{"schema":"public","table_name":"submission","name":"submission_candidate_id_fkey","definition":"FOREIGN KEY (candidate_id) REFERENCES public.candidate(id) MATCH SIMPLE"}},{"id":30078,"constraint_name":"submission_scouter_id_fkey","source_schema":"public","source_table_name":"submission","source_column_name":"scouter_id","target_table_schema":"public","target_table_name":"scouter","target_column_name":"id"}]}]`

	var sourceTables []objects.Table
	err := json.Unmarshal([]byte(jsonStrData), &sourceTables)
	assert.NoError(t, err)

	rs := tables.BuildGenerateModelInputs(sourceTables, nil, nil)

	for _, r := range rs {
		assert.Equal(t, 2, len(r.Relations))
	}
}

func TestAttachActionAndIndex(t *testing.T) {
	jsonStrData := `[{"id":29072,"schema":"public","name":"candidate","rls_enabled":true,"rls_forced":false,"replica_identity":"DEFAULT","bytes":32768,"size":"32 kB","live_rows_estimate":2,"dead_rows_estimate":0,"comment":"list of candidate","columns":[{"table_id":29072,"schema":"public","table":"candidate","id":"29072.1","ordinal_position":1,"name":"id","default_value":null,"data_type":"bigint","format":"int8","is_identity":true,"identity_generation":"BY DEFAULT","is_generated":false,"is_nullable":false,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29072,"schema":"public","table":"candidate","id":"29072.2","ordinal_position":2,"name":"name","default_value":null,"data_type":"character varying","format":"varchar","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29072,"schema":"public","table":"candidate","id":"29072.3","ordinal_position":3,"name":"batch","default_value":null,"data_type":"bigint","format":"int8","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29072,"schema":"public","table":"candidate","id":"29072.4","ordinal_position":4,"name":"created_at","default_value":"now()","data_type":"timestamp with time zone","format":"timestamptz","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null}],"primary_keys":[{"schema":"public","table_name":"candidate","name":"id","table_id":29072}],"relationships":[{"id":29242,"constraint_name":"submission_candidate_id_fkey","source_schema":"public","source_table_name":"submission","source_column_name":"candidate_id","target_table_schema":"public","target_table_name":"candidate","target_column_name":"id"}]},{"id":29079,"schema":"public","name":"scouter","rls_enabled":true,"rls_forced":false,"replica_identity":"DEFAULT","bytes":16384,"size":"16 kB","live_rows_estimate":0,"dead_rows_estimate":0,"comment":"scouter list","columns":[{"table_id":29079,"schema":"public","table":"scouter","id":"29079.1","ordinal_position":1,"name":"id","default_value":null,"data_type":"bigint","format":"int8","is_identity":true,"identity_generation":"BY DEFAULT","is_generated":false,"is_nullable":false,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29079,"schema":"public","table":"scouter","id":"29079.2","ordinal_position":2,"name":"name","default_value":null,"data_type":"character varying","format":"varchar","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29079,"schema":"public","table":"scouter","id":"29079.3","ordinal_position":3,"name":"email","default_value":null,"data_type":"character varying","format":"varchar","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29079,"schema":"public","table":"scouter","id":"29079.4","ordinal_position":4,"name":"created_at","default_value":"now()","data_type":"timestamp with time zone","format":"timestamptz","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null}],"primary_keys":[{"schema":"public","table_name":"scouter","name":"id","table_id":29079}],"relationships":[{"id":30078,"constraint_name":"submission_scouter_id_fkey","source_schema":"public","source_table_name":"submission","source_column_name":"scouter_id","target_table_schema":"public","target_table_name":"scouter","target_column_name":"id"}]},{"id":29086,"schema":"public","name":"submission","rls_enabled":true,"rls_forced":false,"replica_identity":"DEFAULT","bytes":16384,"size":"16 kB","live_rows_estimate":0,"dead_rows_estimate":0,"comment":null,"columns":[{"table_id":29086,"schema":"public","table":"submission","id":"29086.1","ordinal_position":1,"name":"id","default_value":null,"data_type":"bigint","format":"int8","is_identity":true,"identity_generation":"BY DEFAULT","is_generated":false,"is_nullable":false,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29086,"schema":"public","table":"submission","id":"29086.2","ordinal_position":2,"name":"scouter_id","default_value":null,"data_type":"bigint","format":"int8","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29086,"schema":"public","table":"submission","id":"29086.3","ordinal_position":3,"name":"candidate_id","default_value":null,"data_type":"bigint","format":"int8","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29086,"schema":"public","table":"submission","id":"29086.4","ordinal_position":4,"name":"score","default_value":null,"data_type":"real","format":"float4","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29086,"schema":"public","table":"submission","id":"29086.5","ordinal_position":5,"name":"note","default_value":null,"data_type":"text","format":"text","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null},{"table_id":29086,"schema":"public","table":"submission","id":"29086.6","ordinal_position":6,"name":"created_at","default_value":"now()","data_type":"timestamp with time zone","format":"timestamptz","is_identity":false,"identity_generation":null,"is_generated":false,"is_nullable":true,"is_updatable":true,"is_unique":false,"enums":[],"check":null,"comment":null}],"primary_keys":[{"schema":"public","table_name":"submission","name":"id","table_id":29086}],"relationships":[{"id":29242,"constraint_name":"submission_candidate_id_fkey","source_schema":"public","source_table_name":"submission","source_column_name":"candidate_id","target_table_schema":"public","target_table_name":"candidate","target_column_name":"id","action":{"id":29242,"constraint_name":"submission_candidate_id_fkey","deletion_action":"NO ACTION","update_action":"NO ACTION","source_id":29086,"source_schema":"public","source_table":"submission","source_columns":"candidate_id","target_id":29072,"target_schema":"public","target_table":"candidate","target_columns":"id"},"index":{"schema":"public","table_name":"submission","name":"submission_candidate_id_fkey","definition":"FOREIGN KEY (candidate_id) REFERENCES public.candidate(id) MATCH SIMPLE"}},{"id":30078,"constraint_name":"submission_scouter_id_fkey","source_schema":"public","source_table_name":"submission","source_column_name":"scouter_id","target_table_schema":"public","target_table_name":"scouter","target_column_name":"id"}]}]`

	var sourceTables []objects.Table
	err := json.Unmarshal([]byte(jsonStrData), &sourceTables)
	assert.NoError(t, err)

	rs := tables.BuildGenerateModelInputs(sourceTables, nil, nil)
	tbls := make([]objects.Table, 0)
	indexes := make([]objects.Index, 0)
	actions := make([]objects.TablesRelationshipAction, 0)
	for _, r := range rs {
		tbls = append(tbls, r.Table)

		for _, rel := range r.Relations {
			if rel.Action != nil {
				actions = append(actions, *rel.Action)
			}

			if rel.Index != nil {
				indexes = append(indexes, *rel.Index)
			}
		}
	}

	tables.AttachIndexAndAction(tbls, indexes, actions)
}
