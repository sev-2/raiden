package state

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/builder"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/require"
)

type metadataOnly struct{}

type taggedMetadata struct {
	Metadata string `schema:"private" rlsEnable:"false" rlsForced:"true"`
}

type aclModel struct {
	Metadata string `schema:"custom" rlsEnable:"false" rlsForced:"false"`
	ID       int    `column:"name:id;type:integer;primaryKey;nullable:false"`
	Acl      raiden.Acl
}

type pointerAclModel struct {
	Acl *raiden.Acl
}

var valueConfigureCounter int

func (m *aclModel) ConfigureAcl() {
	valueConfigureCounter++
	m.Acl.Enable().Forced()
	m.Acl.Define(
		raiden.Rule("select_data").For("member").To(raiden.CommandSelect).Using(builder.True),
	)
}

func (m *pointerAclModel) ConfigureAcl() {
	configureCounter++
	if m.Acl == nil {
		m.Acl = &raiden.Acl{}
	}
	m.Acl.Enable()
}

var configureCounter int

func TestApplyMetadataVariants(t *testing.T) {
	// Defaults without tags use public schema and enable RLS
	tbDefault := newTableBuilder(metadataOnly{}, nil, nil)
	require.Equal(t, "public", tbDefault.item.Table.Schema)
	require.True(t, tbDefault.item.Table.RLSEnabled)
	require.False(t, tbDefault.item.Table.RLSForced)

	// With metadata tags values should follow tags
	tbTagged := newTableBuilder(taggedMetadata{}, nil, nil)
	require.Equal(t, "private", tbTagged.item.Table.Schema)
	require.False(t, tbTagged.item.Table.RLSEnabled)
	require.True(t, tbTagged.item.Table.RLSForced)

	// When state exists metadata should respect persisted values
	persisted := TableState{Table: objects.Table{Schema: "persisted", RLSEnabled: false, RLSForced: true}}
	tbState := newTableBuilder(metadataOnly{}, nil, &persisted)
	require.Equal(t, "persisted", tbState.item.Table.Schema)
	require.False(t, tbState.item.Table.RLSEnabled)
	require.True(t, tbState.item.Table.RLSForced)
}

func TestApplyPoliciesWithExisting(t *testing.T) {
	statePolicy := objects.Policy{Name: "select_data", Roles: []string{"old"}, Definition: "old"}
	tableState := TableState{
		Table: objects.Table{
			Name:       "acl_model",
			Schema:     "custom",
			RLSEnabled: false,
			RLSForced:  false,
		},
		Policies: []objects.Policy{statePolicy},
	}

	tb := newTableBuilder(&aclModel{}, nil, &tableState)
	tb.processFields()
	tb.applyPolicies()
	result := tb.finish()

	require.True(t, result.Table.RLSEnabled)
	require.True(t, result.Table.RLSForced)
	require.Len(t, result.ExtractedPolicies.Existing, 1)

	updated := result.ExtractedPolicies.Existing[0]
	require.Contains(t, updated.Roles, "member")
	require.Equal(t, builder.True.String(), updated.Definition)
	require.NotContains(t, tb.existingPolicies, statePolicy.Name)
}

func TestGetAclVariants(t *testing.T) {
	valueConfigureCounter = 0
	configureCounter = 0

	// Value field
	model := &aclModel{}
	aclInstance := getAcl(model)
	require.NotNil(t, aclInstance)
	require.Equal(t, 1, valueConfigureCounter)
	// Second call does not reconfigure
	again := getAcl(model)
	require.Equal(t, aclInstance, again)
	require.Equal(t, 0, configureCounter)

	// Pointer field with nil value should be initialised
	configureCounter = 0
	pointerModel := &pointerAclModel{}
	aclPtr := getAcl(pointerModel)
	require.NotNil(t, aclPtr)
	require.Equal(t, 1, configureCounter)
	require.NotNil(t, pointerModel.Acl)

	// Model without ACL should return nil
	require.Nil(t, getAcl(struct{}{}))
}

func TestApplyPoliciesWithNoAcl(t *testing.T) {
	tb := newTableBuilder(struct {
		ID int `column:"name:id;type:integer"`
	}{}, nil, nil)
	tb.processFields()
	require.NotPanics(t, func() { tb.applyPolicies() })
	require.Empty(t, tb.item.ExtractedPolicies.New)
}
