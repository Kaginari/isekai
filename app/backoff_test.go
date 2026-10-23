package app

import (
	"time"

	"github.com/Kaginari/isekai/provider"
)

func init() {
	provider.Backoff = func(int) time.Duration { return time.Millisecond }
	ladder = []time.Duration{time.Millisecond, time.Millisecond}
}
