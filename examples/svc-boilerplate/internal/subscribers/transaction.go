package subscribers

import (
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

func (s *RideHailingSubscriber) Subscription() string {
	return "transaction.ride-hailing.new-sub"
}

func (s *RideHailingSubscriber) Consume(ctx raiden.SubscriberContext, message raiden.SubscriberMessage) error {
	raiden.Info("ride hailing new status listener executed", "data", string(message.Data))
	return nil
}
