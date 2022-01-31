package jsonlog

import (
	"context"
	"time"
)

type Time func() time.Time

type Hook func(ctx context.Context, event *Event)