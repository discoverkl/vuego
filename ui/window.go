package ui

type Window interface {
	// Load(url string) error
	// Bounds() (Bounds, error)
	// SetBounds(Bounds) error
	Bind(b Bindings) error
	Eval(js string) Value
	Done() <-chan struct{}
	Close() error
}
