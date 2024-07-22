package generator_test

import (
	"os"
	"path/filepath"
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

func TestGenerateRpcRegister(t *testing.T) {
	dir, err := os.MkdirTemp("", "rpc_register")
	assert.NoError(t, err)

	rpcPath := filepath.Join(dir, "internal")
	err1 := utils.CreateFolder(rpcPath)
	assert.NoError(t, err1)

	err2 := generator.GenerateRpcRegister(dir, "test", generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)
	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/bootstrap"))
	assert.FileExists(t, dir+"/internal/bootstrap/rpc.go")
}
