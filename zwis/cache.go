package zwis

import (
	"context"
	"time"
)

/*
Cache Interface will contain the following key parameters:

Set()
Get()
Delete()
Flush()

*/

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string) (interface{}, bool)
	Delete(ctx context.Context, key string) error
	Flush(ctx context.Context) error
}
