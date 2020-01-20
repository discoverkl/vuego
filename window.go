package vuego

type Window interface {
	// Load(url string) error
	// Bounds() (Bounds, error)
	// SetBounds(Bounds) error
	Bind(name string, f interface{}) error
	Eval(js string) Value
	Done() <-chan struct{}
	Close() error
}