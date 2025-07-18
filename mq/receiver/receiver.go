package receiver

import "context"

type Receiver interface {
	ReceiveOrder(ctx context.Context)
}
