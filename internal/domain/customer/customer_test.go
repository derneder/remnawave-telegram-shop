package customer

import "testing"
import "time"

func TestCustomerFields(t *testing.T) {
	now := time.Now()
	link := "link"
	c := Customer{
		ID:               1,
		TelegramID:       2,
		ExpireAt:         &now,
		SubscriptionLink: &link,
		Language:         "en",
		Balance:          3.5,
	}
	if c.TelegramID != 2 || *c.ExpireAt != now || c.Balance != 3.5 {
		t.Fatalf("unexpected values: %#v", c)
	}
}
