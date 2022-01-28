package jsonlog

import "context"

type Hook func(ctx context.Context, event *Event)

var hooks []Hook

func RegisterHook(h Hook) {
	hooks = append(hooks, h)
}
