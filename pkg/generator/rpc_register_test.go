package generator_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestWalkRpcDir(t *testing.T) {

	testPath, err := utils.GetAbsolutePath("/testdata")
	assert.NoError(t, err)

	rs, err := generator.WalkScanRpc(testPath)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(rs))
	assert.Equal(t, "GetVoteBy", rs[0])
}
