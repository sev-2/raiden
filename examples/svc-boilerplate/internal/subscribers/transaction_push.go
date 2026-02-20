package subscribers

import (
	"encoding/json"

	"github.com/sev-2/raiden"
)

type RideHailingPushSubscriber struct {
	raiden.SubscriberBase
}

func (s *RideHailingPushSubscriber) Name() string {
	return "RideHailingPush New"
}

func (s *RideHailingPushSubscriber) Provider() raiden.PubSubProviderType {
	return raiden.PubSubProviderGoogle
}

func (s *RideHailingPushSubscriber) Topic() string {
	return "transaction.ride-hailing.new"
}

func (s *RideHailingPushSubscriber) Subscription() string {
	return "transaction.ride-hailing.paid"
}

func (s *RideHailingPushSubscriber) SubscriptionType() raiden.SubscriptionType {
	return raiden.SubscriptionTypePush
}

func (s *RideHailingPushSubscriber) PushEndpoint() string {
	return "/transaction-ride"
}

func (s *RideHailingPushSubscriber) Consume(ctx raiden.SubscriberContext, message raiden.SubscriberMessage) error {
	dByte, _ := json.MarshalIndent(message, " ", " ")
	raiden.Info("push subscription message", "data", string(dByte))
	return nil
}
