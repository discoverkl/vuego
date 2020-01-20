package vuego

import (
	"encoding/json"
	"log"
)

// Function wraps a js callback function.
type Function struct {
	BindingName string `json:"bindingName"`
	Seq         int    `json:"seq"`

	jsc *jsClient
}

// close is called by page automatically
func (c *Function) close() {
	_, err := c.jsc.send("Vuego.closeCallback", h{"name": c.BindingName, "seq": c.Seq}, false)
	if err != nil {
		log.Println("close callback failed:", err)
	}
}

// Call method.
func (c *Function) Call(args ...interface{}) Value {
	_, err := json.Marshal(args)
	if err != nil {
		return value{err: err}
	}
	v, err := c.jsc.send("Vuego.callback", h{"name": c.BindingName, "seq": c.Seq, "args": args}, true)
	return value{err: err, raw: v}
}
