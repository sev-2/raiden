package raiden_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type pubsubHandler struct {
	raiden.SubscriberBase
}

func (*pubsubHandler) Provider() raiden.PubSubProviderType {
	return raiden.PubSubProviderGoogle
}

func TestPubsubInstance(t *testing.T) {
	pubsub := raiden.NewPubsub(nil, nil)
	assert.NotNil(t, pubsub)
}

func TestNewGooglePubsub(t *testing.T) {
	config := loadConfig()
	config.GoogleProjectId = "x"
	config.GoogleSaPath = "/sa.json"

	mockProvider := mock.MockProvider{
		PublishFn: func(ctx context.Context, topic string, message []byte) error {
			return nil
		},
		StartListenFn: func(handler []raiden.SubscriberHandler) error {
			return nil
		},
		StopListenFn: func() error {
			return nil
		},
	}

	pubsub := raiden.PubSubManager{}

	pubsub.SetConfig(config)
	pubsub.SetProvider(raiden.PubSubProviderGoogle, &mockProvider)

	handler := &pubsubHandler{}

	assert.Equal(t, true, handler.AutoAck())
	assert.Equal(t, "unknown", handler.Name())
	assert.Equal(t, "", handler.Topic())
	assert.Error(t, handler.Consume(nil, nil))

	// assert register
	pubsub.Register(handler)
	assert.Equal(t, 1, pubsub.GetHandlerCount())

	// assert publish
	err := pubsub.Publish(context.Background(), raiden.PubSubProviderGoogle, "test", []byte("{\"message\":\"hello\"}"))
	assert.NoError(t, err)

	mockProvider.PublishFn = func(ctx context.Context, topic string, message []byte) error { return errors.New("test error") }
	err = pubsub.Publish(context.Background(), raiden.PubSubProviderGoogle, "test", []byte("{\"message\":\"hello\"}"))
	assert.Error(t, err)

	pubsub.Listen()

	mockProvider.PublishFn = func(ctx context.Context, topic string, message []byte) error { return nil }
	err = pubsub.Publish(context.Background(), raiden.PubSubProviderUnknown, "test", []byte("{\"message\":\"hello\"}"))
	assert.Error(t, err)

	mockProvider.StopListenFn = func() error { return errors.New("fail stop") }
	pubsub.Listen()
}
