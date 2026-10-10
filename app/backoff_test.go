package app

import (
	"time"

	"github.com/Kaginari/isekai/provider"
)

func init() { provider.Backoff = func(int) time.Duration { return time.Millisecond } }
