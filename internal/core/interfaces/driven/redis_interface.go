package driveninterfaces

import (
	"context"
	"time"
)

type RedisDrivenInterface interface {
	Set(ctx context.Context, key string, value interface{}, duration time.Duration) error
}
