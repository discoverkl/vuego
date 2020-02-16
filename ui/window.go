package ui

import "time"

type Window interface {
	// Load(url string) error
	// Bounds() (Bounds, error)
	// SetBounds(Bounds) error
	Bind(b Bindings) error
	Open() error
	SetExitDelay(d time.Duration)
	Eval(js string) Value
	Done() <-chan struct{}
	Close() error
}
