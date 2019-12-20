package one

import (
	"context"
	"log"
	"sync/atomic"
	"time"
)

var dev = InDevMode("one")

type debounceArgs struct {
	immediate bool
}

type DebounceOption func(args *debounceArgs)

func DebounceImmediate() DebounceOption {
	return func(args *debounceArgs) {
		args.immediate = true
	}
}

func Debounce(ctx context.Context, fn func(), wait time.Duration, ops ...DebounceOption) func() {
	args := &debounceArgs{}
	for _, op := range ops {
		op(args)
	}

	if args.immediate {
		panic("not implemented: immediate")
	}

	var hotLevel int64
	ticker := time.NewTicker(wait / 2)
	var lastHit time.Time
	go func() {
		defer ticker.Stop()
		if dev {
			log.Println("[Debounce Ticker] Enter")
			defer log.Println("[Debounce Ticker] Leave")
		}
		for {
			select {
			case <-ticker.C:
				level := atomic.LoadInt64(&hotLevel)
				switch level {
				case 0:
					break
				case 1:
					atomic.StoreInt64(&hotLevel, 0)
					if dev {
						realWait := time.Now().Sub(lastHit)
						log.Printf("real wait: %.3fs", realWait.Seconds())
					}
					go fn()
				case 2:
					atomic.StoreInt64(&hotLevel, 1)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return func() {
		if dev {
			lastHit = time.Now()
		}
		atomic.StoreInt64(&hotLevel, 2)
	}
}
