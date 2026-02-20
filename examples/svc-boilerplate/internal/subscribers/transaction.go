package subscribers

import (
	"encoding/json"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/examples/svc-boilerplate/internal/models"
	"github.com/sev-2/raiden/pkg/db"
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
	var book models.Book
	if err := json.Unmarshal(message.Data, &book); err != nil {
		return err
	}

	activityEntry := new(any)
	activityLog := models.Activity{ItemId: book.Id, Activity: "create transaction at " + time.Now().String()}
	if err := db.NewQuery(nil).From(models.Activity{}).Insert(activityLog, &activityEntry); err != nil {
		raiden.Error("Insert err : ", err)
		return err
	}

	raiden.Info("activity : ", activityEntry)
	raiden.Info("ride haling new status listener executed", "data", string(message.Data))
	return nil
}
