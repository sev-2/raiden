package raiden_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalAclTag(t *testing.T) {
	// Test case 1
	tag := `read:"role1,role2",write:"role3,role4",readUsing:" using1",writeCheck:"check",writeUsing:"using2"` // tag string
	aclTag := raiden.UnmarshalAclTag(tag)                                                                      // call UnmarshalAclTag function
	assert.NotNil(t, aclTag)                                                                                   // check if aclTag is not nil
	assert.Equal(t, []string{"role1", "role2"}, aclTag.Read.Roles)                                             // check if aclTag.Read.Roles is equal to []string{"role1", "role2"}
	assert.Equal(t, []string{"role3", "role4"}, aclTag.Write.Roles)                                            // check if aclTag.Write.Roles is equal to []string{"role3", "role4"}
	assert.Equal(t, " using1", aclTag.Read.Using)                                                              // check if aclTag.Read.Using is equal to " using1"
	assert.NotNil(t, aclTag.Write.Check)                                                                       // check if aclTag.Write.Check is not nil
	assert.Equal(t, "check", *aclTag.Write.Check)                                                              // check if *aclTag.Write.Check is equal to "check"
	assert.Equal(t, "using2", aclTag.Write.Using)                                                              // check if aclTag.Write.Using is equal to "using2"

	// Test case 2
	tag = `read:"role1,role2",write:"role3,role4",readUsing:" using1",writeCheck:"check"` // tag string
	aclTag = raiden.UnmarshalAclTag(tag)                                                  // call UnmarshalAclTag function
	assert.NotNil(t, aclTag)                                                              // check if aclTag is not nil
	assert.Equal(t, []string{"role1", "role2"}, aclTag.Read.Roles)                        // check if aclTag.Read.Roles is equal to []string{"role1", "role2"}
	assert.Equal(t, []string{"role3", "role4"}, aclTag.Write.Roles)                       // check if aclTag.Write.Roles is equal to []string{"role3", "role4"}
	assert.Equal(t, " using1", aclTag.Read.Using)                                         // check if aclTag.Read.Using is equal to " using1"
	assert.NotNil(t, aclTag.Write.Check)                                                  // check if aclTag.Write.Check is not nil
	assert.Equal(t, "check", *aclTag.Write.Check)                                         // check if *aclTag.Write.Check is equal to "check"
	assert.Empty(t, aclTag.Write.Using)                                                   // check if aclTag.Write.Using is empty

	// Test case 3

	tag = `read:"role1,role2",write:"role3,role4",readUsing:" using1"` // tag string
	aclTag = raiden.UnmarshalAclTag(tag)                               // call UnmarshalAclTag function
	assert.NotNil(t, aclTag)                                           // check if aclTag is not nil
	assert.Equal(t, []string{"role1", "role2"}, aclTag.Read.Roles)     // check if aclTag.Read.Roles is equal to []string{"role1", "role2"}
	assert.Equal(t, []string{"role3", "role4"}, aclTag.Write.Roles)    // check if aclTag.Write.Roles is equal to []string{"role3", "role4"}
	assert.Equal(t, " using1", aclTag.Read.Using)                      // check if aclTag.Read.Using is equal to " using1"
	assert.Empty(t, aclTag.Write.Check)                                // check if aclTag.Write.Check is empty
	assert.Empty(t, aclTag.Write.Using)                                // check if aclTag.Write.Using is empty

	// Test case 4
	tag = `read:"role1,role2",write:"role3,role4"`                  // tag string
	aclTag = raiden.UnmarshalAclTag(tag)                            // call UnmarshalAclTag function
	assert.NotNil(t, aclTag)                                        // check if aclTag is not nil
	assert.Equal(t, []string{"role1", "role2"}, aclTag.Read.Roles)  // check if aclTag.Read.Roles is equal to []string{"role1", "role2"}
	assert.Equal(t, []string{"role3", "role4"}, aclTag.Write.Roles) // check if aclTag.Write.Roles is equal to []string{"role3", "role4"}
	assert.Empty(t, aclTag.Read.Using)                              // check if aclTag.Read.Using is empty
	assert.Empty(t, aclTag.Write.Check)                             // check if aclTag.Write.Check is empty
	assert.Empty(t, aclTag.Write.Using)                             // check if aclTag.Write.Using is empty

	// Test case 5
	tag = `read:"role1,role2"`                                     // tag string
	aclTag = raiden.UnmarshalAclTag(tag)                           // call UnmarshalAclTag function
	assert.NotNil(t, aclTag)                                       // check if aclTag is not nil
	assert.Equal(t, []string{"role1", "role2"}, aclTag.Read.Roles) // check if aclTag.Read.Roles is equal to []string{"role1", "role2"}
	assert.Empty(t, aclTag.Write.Roles)                            // check if aclTag.Write.Roles is empty
	assert.Empty(t, aclTag.Read.Using)                             // check if aclTag.Read.Using is empty
	assert.Empty(t, aclTag.Write.Check)                            // check if aclTag.Write.Check is empty
	assert.Empty(t, aclTag.Write.Using)                            // check if aclTag.Write.Using is empty

	// Test case 6
	tag = `write:"role3,role4"`                                     // tag string
	aclTag = raiden.UnmarshalAclTag(tag)                            // call UnmarshalAclTag function
	assert.NotNil(t, aclTag)                                        // check if aclTag is not nil
	assert.Empty(t, aclTag.Read.Roles)                              // check if aclTag.Read.Roles is empty
	assert.Equal(t, []string{"role3", "role4"}, aclTag.Write.Roles) // check if aclTag.Write.Roles is equal to []string{"role3", "role4"}
	assert.Empty(t, aclTag.Read.Using)                              // check if aclTag.Read.Using is empty
	assert.Empty(t, aclTag.Write.Check)                             // check if aclTag.Write.Check is empty
	assert.Empty(t, aclTag.Write.Using)                             // check if aclTag.Write.Using is empty

	// Test case 7
	tag = `readUsing:" using1"`                   // tag string
	aclTag = raiden.UnmarshalAclTag(tag)          // call UnmarshalAclTag function
	assert.NotNil(t, aclTag)                      // check if aclTag is not nil
	assert.Empty(t, aclTag.Read.Roles)            // check if aclTag.Read.Roles is empty
	assert.Empty(t, aclTag.Write.Roles)           // check if aclTag.Write.Roles is empty
	assert.Equal(t, " using1", aclTag.Read.Using) // check if aclTag.Read.Using is equal to " using1"
	assert.Empty(t, aclTag.Write.Check)           // check if aclTag.Write.Check is empty
	assert.Empty(t, aclTag.Write.Using)           // check if aclTag.Write.Using is empty

	// Test case 8
	tag = `writeCheck:"check"`                    // tag string
	aclTag = raiden.UnmarshalAclTag(tag)          // call UnmarshalAclTag function
	assert.NotNil(t, aclTag)                      // check if aclTag is not nil
	assert.Empty(t, aclTag.Read.Roles)            // check if aclTag.Read.Roles is empty
	assert.Empty(t, aclTag.Write.Roles)           // check if aclTag.Write.Roles is empty
	assert.Empty(t, aclTag.Read.Using)            // check if aclTag.Read.Using is empty
	assert.NotNil(t, aclTag.Write.Check)          // check if aclTag.Write.Check is not nil
	assert.Equal(t, "check", *aclTag.Write.Check) // check if *aclTag.Write.Check is equal to "check"
	assert.Empty(t, aclTag.Write.Using)           // check if aclTag.Write.Using is empty

	// Test case 9
	tag = `write Using:"using2"`         // tag string
	aclTag = raiden.UnmarshalAclTag(tag) // call UnmarshalAclTag function
	assert.NotNil(t, aclTag)             // check if aclTag is not nil
	assert.Empty(t, aclTag.Read.Roles)   // check if aclTag.Read.Roles is empty
	assert.Empty(t, aclTag.Write.Roles)  // check if aclTag.Write.Roles is empty
	assert.Empty(t, aclTag.Read.Using)   // check if aclTag.Read.Using is empty
	assert.Empty(t, aclTag.Write.Check)  // check if aclTag.Write.Check is empty
	assert.Empty(t, aclTag.Write.Using)  // check if aclTag.Write.Using is empty

	// Test case 10
	tag = `read:"role1,role2",write:"role3,role4",readUsing:" using1",writeCheck:"check",writeUsing:"using2"` // tag string
	aclTag = raiden.UnmarshalAclTag(tag)                                                                      // call UnmarshalAclTag function
	assert.NotNil(t, aclTag)                                                                                  // check if aclTag is not nil
	assert.Equal(t, []string{"role1", "role2"}, aclTag.Read.Roles)                                            // check if aclTag.Read.Roles is equal to []string{"role1", "role2"}
	assert.Equal(t, []string{"role3", "role4"}, aclTag.Write.Roles)                                           // check if aclTag.Write.Roles is equal to []string{"role3", "role4"}
	assert.Equal(t, " using1", aclTag.Read.Using)                                                             // check if aclTag.Read.Using is equal to " using1"
	assert.NotNil(t, aclTag.Write.Check)                                                                      // check if aclTag.Write.Check is not nil
	assert.Equal(t, "check", *aclTag.Write.Check)                                                             // check if *aclTag.Write.Check is equal to "check"
	assert.Equal(t, "using2", aclTag.Write.Using)                                                             // check if aclTag.Write.Using is equal to "using2"
}
