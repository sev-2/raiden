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

import (
	"errors"

	"cloud.google.com/go/pubsub"
	"github.com/sev-2/raiden"
)

type RideHailingSubscriber struct {
	raiden.SubscriberBase
}

func (s *RideHailingSubscriber) Name() string {
	return "RideHailing New"
}

func (s *RideHailingSubscriber) Provider() raiden.PubSubProviderType {
	return raiden.PubSubProviderGoogle
}

func (s *RideHailingSubscriber) Subscripbtion() string {
	return "transaction.ride-hailing.new-sub"
}

func (s *RideHailingSubscriber) Consume(ctx raiden.SubscriberContext, message any) error {
	msg, valid := message.(*pubsub.Message)
	if !valid {
		return errors.New("invalid google pubsub message")
	}

	raiden.Info("ride haling new status listener executed", "data", string(msg.Data))
	return nil
}
`
	_, err4 := sampleSubscriberFile.WriteString(configContent)
	assert.NoError(t, err4)
	sampleSubscriberFile.Close()

	foundFiles, err5 := generator.WalkScanSubscriber(dir + "/internal/subscribers")
	assert.NoError(t, err5)
	assert.NotEmpty(t, foundFiles)

	err5 = generator.GenerateSubscriberRegister(dir, "test", generator.GenerateFn(generator.Generate))
	assert.NoError(t, err5)
}

func TestGenerateSubscriberRegister_Empty(t *testing.T) {
	dir, err := os.MkdirTemp("", "subscriber_register")
	assert.NoError(t, err)

	subscriberPath := filepath.Join(dir, "internal")
	err1 := utils.CreateFolder(subscriberPath)
	assert.NoError(t, err1)

	err2 := generator.GenerateSubscriberRegister(dir, "test", generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)
	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/bootstrap"))
	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/subscribers"))

	foundFiles, err3 := generator.WalkScanSubscriber(dir + "/internal/subscribers")
	assert.NoError(t, err3)
	assert.Empty(t, foundFiles)
}
