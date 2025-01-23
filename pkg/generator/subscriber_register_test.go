package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGenerateSubscriberRegister(t *testing.T) {
	dir, err := os.MkdirTemp("", "subscriber_register")
	assert.NoError(t, err)

	subscriberPath := filepath.Join(dir, "internal")
	err1 := utils.CreateFolder(subscriberPath)
	assert.NoError(t, err1)

	err2 := generator.GenerateSubscriberRegister(dir, "test", generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)
	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/bootstrap"))
	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/subscribers"))

	sampleSubscriberFile, err3 := utils.CreateFile(dir+"/internal/subscribers/sample_subscriber.go", true)
	assert.NoError(t, err3)

	configContent := `
package subscribers


type NiceSubscriber struct {
	raiden.SubscriberBase
}

func (j *NiceSubscriber) Name() string {
	return "some-nice-subscriber"
}
`
	_, err4 := sampleSubscriberFile.WriteString(configContent)
	assert.NoError(t, err4)
	sampleSubscriberFile.Close()

	foundFiles, err5 := generator.WalkScanSubscriber(dir + "/internal/subscribers")
	assert.NoError(t, err5)
	assert.NotEmpty(t, foundFiles)
}
