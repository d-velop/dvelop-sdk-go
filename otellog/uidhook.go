package otellog

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

type UidFromContextFn func(ctx context.Context) (string, error)

func AddUserIdToLogEvents(uidFn UidFromContextFn) Hook {

	return func(ctx context.Context, e *Event) {
		uid, _ := uidFn(ctx)
		if uid != "" {
			hash := sha256.New()
			hash.Write([]byte(uid))
			sum := hash.Sum(nil)
			uidHash := hex.EncodeToString(sum)
			e.UserIdHash = uidHash
		}
	}

}
