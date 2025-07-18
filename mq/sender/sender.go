package sender

import "context"

type Sender interface {
	SendOrder(ctx context.Context)
}
