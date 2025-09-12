package consumer

import (
	"context"
)

type Broker interface {
	Start(ctx context.Context) error
	Stop() error
}
