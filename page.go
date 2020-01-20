package vuego

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"golang.org/x/net/websocket"
)

// Page of a javascript client.
type Page interface {
	Bind(name string, f interface{}) error
	Eval(js string) Value
	Done() <-chan struct{}
	Ready() error	// nofity server ready ( all functions binded )
}

type page struct {
	jsc  *jsClient
	done chan struct{}
}

func newPage(ws *websocket.Conn) (Page, error) {
	jsc, err := newJSClient(ws)
	if err != nil {
		return nil, err
	}
	ui := &page{
		jsc:  jsc,
		done: make(chan struct{}),
	}
	return ui, nil
}

func (c *page) Bind(name string, f interface{}) error {
	if err := checkBindFunc(f); err != nil {
		return err
	}
	v := reflect.ValueOf(f)
	return c.jsc.bind(name, func(raw []json.RawMessage) (interface{}, error) {
		// Vuego.call -> here(do the real call) -> eval for promise
		if len(raw) != v.Type().NumIn() {
			return nil, fmt.Errorf("function arguments mismatch")
		}
		args := []reflect.Value{}

		// TODO: argumets rewrite
		functionType := reflect.TypeOf((**Function)(nil))
		for i := range raw {
			arg := reflect.New(v.Type().In(i))
			if err := json.Unmarshal(raw[i], arg.Interface()); err != nil {
				return nil, err
			}
			if arg.Type() == functionType {
				fn := arg.Elem().Interface().(*Function)
				fn.jsc = c.jsc
				defer fn.close()
			}
			args = append(args, arg.Elem())
		}

		errorType := reflect.TypeOf((*error)(nil)).Elem()
		res := v.Call(args)
		switch len(res) {
		case 0:
			// no return value
			return nil, nil
		case 1:
			// return value or error
			if res[0].Type().Implements(errorType) {
				if res[0].Interface() != nil {
					return nil, res[0].Interface().(error)
				}
				return nil, nil
			}
			return res[0].Interface(), nil
		case 2:
			// first one is value, second is error
			if !res[1].Type().Implements(errorType) {
				return nil, errors.New("second return value must be an error")
			}
			if res[1].Interface() == nil {
				return res[0].Interface(), nil
			}
			return res[0].Interface(), res[1].Interface().(error)
		default:
			return nil, errors.New("unexpected number of return values")
		}
	})
}

func (c *page) Eval(js string) Value {
	v, err := c.jsc.eval(js)
	return value{err: err, raw: v}
}

func (c *page) Done() <-chan struct{} {
	return c.done
}

func (c *page) Ready() error {
	return c.jsc.ready()
}

func checkBindFunc(f interface{}) error {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return fmt.Errorf("f should be a function")
	}
	if n := v.Type().NumOut(); n > 2 {
		return fmt.Errorf("too many return values")
	}
	return nil
}
