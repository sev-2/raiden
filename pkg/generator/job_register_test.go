package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGenerateJobRegister(t *testing.T) {
	dir, err := os.MkdirTemp("", "job_register")
	assert.NoError(t, err)

	jobPath := filepath.Join(dir, "internal")
	err1 := utils.CreateFolder(jobPath)
	assert.NoError(t, err1)

	err2 := generator.GenerateJobRegister(dir, "test", generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)
	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/bootstrap"))
	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/jobs"))

	sampleJobFile, err3 := utils.CreateFile(dir+"/internal/jobs/sample_job.go", true)
	assert.NoError(t, err3)

	configContent := `
package jobs


type NiceJob struct {
	raiden.JobBase
}

func (j *NiceJob) Name() string {
	return "some-nice-job"
}
`
	_, err4 := sampleJobFile.WriteString(configContent)
	assert.NoError(t, err4)
	sampleJobFile.Close()

	foundFiles, err5 := generator.WalkScanJob(dir + "/internal/jobs")
	assert.NoError(t, err5)
	assert.NotEmpty(t, foundFiles)
}
