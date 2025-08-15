package drivenports

import "context"

type PriceCache interface {
	SetLatest(ctx context.Context, symbol string, price float64) error
	SetLatestByEx(ctx context.Context, ex, symbol string, price float64) error
	ZAddTick(ctx context.Context, ex, symbol string, ts int64, price float64) error
	ZTrimBefore(ctx context.Context, ex, symbol string, ts int64) error
}
