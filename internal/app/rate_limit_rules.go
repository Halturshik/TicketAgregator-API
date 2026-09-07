package app

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/ratelimit"
)

var (
	guestSearchPolicy  = ratelimit.Policy{Name: "search:guest", Limit: 10, Window: time.Minute}
	userSearchPolicy   = ratelimit.Policy{Name: "search:user", Limit: 20, Window: time.Minute}
	searchPagePolicy   = ratelimit.Policy{Name: "search:page", Limit: 60, Window: time.Minute}
	publicLookupPolicy = ratelimit.Policy{
		Name: "booking:public_lookup", Limit: 20, Window: time.Minute,
	}
)
