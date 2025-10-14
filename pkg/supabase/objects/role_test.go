package objects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRolesToMap(t *testing.T) {
	// Test with empty roles slice
	emptyRoles := Roles{}
	result := emptyRoles.ToMap()
	assert.Equal(t, 0, len(result))

	// Test with single role
	singleRole := Roles{
		{ID: 1, Name: "role1"},
	}
	result = singleRole.ToMap()
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "role1", result[1].Name)

	// Test with multiple roles
	multipleRoles := Roles{
		{ID: 1, Name: "role1"},
		{ID: 2, Name: "role2"},
		{ID: 3, Name: "role3"},
	}
	result = multipleRoles.ToMap()
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "role1", result[1].Name)
	assert.Equal(t, "role2", result[2].Name)
	assert.Equal(t, "role3", result[3].Name)

	// Test with duplicate IDs (the last one should overwrite previous)
	duplicateRoles := Roles{
		{ID: 1, Name: "first-role"},
		{ID: 2, Name: "second-role"},
		{ID: 1, Name: "overwritten-role"}, // Same ID as first, should overwrite
	}
	result = duplicateRoles.ToMap()
	assert.Equal(t, 2, len(result))                     // Only 2 unique IDs
	assert.Equal(t, "overwritten-role", result[1].Name) // Should have overwritten
	assert.Equal(t, "second-role", result[2].Name)
}

func TestRoleMembershipsGroupByInheritId(t *testing.T) {
	// Test with empty role memberships slice
	emptyMemberships := RoleMemberships{}
	result := emptyMemberships.GroupByInheritId()
	assert.Equal(t, 0, len(result))

	// Test with single role membership
	singleMembership := RoleMemberships{
		{ParentID: 1, ParentRole: "parent1", InheritID: 10, InheritRole: "inherit1"},
	}
	result = singleMembership.GroupByInheritId()
	assert.Equal(t, 1, len(result))
	assert.Equal(t, 1, len(result[10]))
	assert.Equal(t, "parent1", result[10][0].ParentRole)
	assert.Equal(t, "inherit1", result[10][0].InheritRole)

	// Test with multiple role memberships, same inherit ID
	sameInheritId := RoleMemberships{
		{ParentID: 1, ParentRole: "parent1", InheritID: 10, InheritRole: "inherit1"},
		{ParentID: 2, ParentRole: "parent2", InheritID: 10, InheritRole: "inherit1"}, // Same InheritID
		{ParentID: 3, ParentRole: "parent3", InheritID: 10, InheritRole: "inherit1"}, // Same InheritID
	}
	result = sameInheritId.GroupByInheritId()
	assert.Equal(t, 1, len(result))     // Only 1 unique InheritID
	assert.Equal(t, 3, len(result[10])) // 3 items for InheritID 10

	// Test with multiple role memberships, different inherit IDs
	multipleInheritIds := RoleMemberships{
		{ParentID: 1, ParentRole: "parent1", InheritID: 10, InheritRole: "inherit1"},
		{ParentID: 2, ParentRole: "parent2", InheritID: 20, InheritRole: "inherit2"},
		{ParentID: 3, ParentRole: "parent3", InheritID: 10, InheritRole: "inherit1"}, // Same InheritID as first
		{ParentID: 4, ParentRole: "parent4", InheritID: 30, InheritRole: "inherit3"},
	}
	result = multipleInheritIds.GroupByInheritId()
	assert.Equal(t, 3, len(result))     // 3 unique InheritIDs: 10, 20, 30
	assert.Equal(t, 2, len(result[10])) // 2 items for InheritID 10
	assert.Equal(t, 1, len(result[20])) // 1 item for InheritID 20
	assert.Equal(t, 1, len(result[30])) // 1 item for InheritID 30

	// Verify the contents are correct
	assert.Equal(t, "parent1", result[10][0].ParentRole)
	assert.Equal(t, "parent3", result[10][1].ParentRole)
	assert.Equal(t, "parent2", result[20][0].ParentRole)
	assert.Equal(t, "parent4", result[30][0].ParentRole)
}
