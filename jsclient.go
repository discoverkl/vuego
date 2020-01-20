package vuego

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"sync"
	"sync/atomic"

	"golang.org/x/net/websocket"
)

type result struct {
	Value json.RawMessage
	Err   error
}

type bindingFunc func(args []json.RawMessage) (interface{}, error)

type msg struct {
	ID     int             `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	// Result json.RawMessage `json:"result"`
	// Error  json.RawMessage `json:"error"`
}

type retParams struct {
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error"`
}

type callParams struct {
	Name string            `json:"name"`
	Seq  int               `json:"seq"`
	Args []json.RawMessage `json:"args"`
}

type h map[string]interface{}

type jsClient struct {
	sync.Mutex
	id      int32
	pending map[int]chan result
	ws      *websocket.Conn
	binding map[string]bindingFunc
	done chan struct{}
}

func newJSClient(ws *websocket.Conn) (*jsClient, error) {
	p := &jsClient{
		ws:      ws,
		pending: map[int]chan result{},
		binding: map[string]bindingFunc{},
		done: make(chan struct{}),
	}
	go p.readLoop()
	return p, nil
}

func (p *jsClient) readLoop() {
	defer close(p.done)
	for {
		m := msg{}
		if err := websocket.JSON.Receive(p.ws, &m); err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("remote closed")
				return
			}
			log.Println("receive bad message:", err)
			continue
		}
		// log.Printf("[receive] %s", m.Method)

		switch m.Method {
		case "1":
			go func() {
				_, err := p.send("Vuego.call", h{"name": "eval", "args": []string{"1+2"}}, true)
				if err != nil {
					log.Println("send message:", err)
				}
			}()
		case "Vuego.ret":
			ret := retParams{}
			err := json.Unmarshal([]byte(m.Params), &ret)
			if err != nil {
				log.Println("Vuego.ret bad message:", err)
			}

			p.Lock()
			retCh, ok := p.pending[m.ID]
			delete(p.pending, m.ID)
			p.Unlock()

			if !ok {
				var v interface{}
				err = json.Unmarshal(ret.Result, &v)
				valid := (err == nil)
				log.Printf("ignore Vuego.ret %d: valid=%v ret=%v, err=%s", m.ID, valid, v, ret.Error)
				continue
			}

			if ret.Error != "" {
				retCh <- result{Err: errors.New(ret.Error)}
			} else {
				retCh <- result{Value: ret.Result}
			}
		case "Vuego.call":
			call := callParams{}
			err := json.Unmarshal([]byte(m.Params), &call)
			if err != nil {
				log.Println("Vuego.call bad message:", err)
			}

			p.Lock()
			binding, ok := p.binding[call.Name]
			p.Unlock()

			if !ok {
				break
			}

			go func() {
				// jsRet is null or string, jsErr is json value
				var jsRet, jsErr interface{}
				// binding call phrase 2
				if ret, err := binding(call.Args); err != nil {
					jsErr = err.Error()
				} else if _, err = json.Marshal(ret); err != nil {
					jsErr = err.Error()
				} else {
					jsRet = ret
				}
				_, err = p.send("Vuego.ret", h{"name": call.Name, "seq": call.Seq, "result": jsRet, "error": jsErr}, false)
				if err != nil {
					log.Println("binding call phrase 3 failed:", err)
				}
			}()

		default:
			log.Println("unknown method:", m.Method)
		}
	}
}

func (p *jsClient) send(method string, params h, wait bool) (json.RawMessage, error) {
	log.Printf("[send] method %s, wait=%v", method, wait)
	id := atomic.AddInt32(&p.id, 1)
	m := h{"id": int(id), "method": method, "params": params}

	var retCh chan result
	if wait {
		retCh = make(chan result)
		p.Lock()
		p.pending[int(id)] = retCh
		p.Unlock()
	}

	err := websocket.JSON.Send(p.ws, m)
	if err != nil {
		// TODO: remove item in p.pending
		return nil, err
	}

	if !wait {
		return nil, nil
	}
	ret := <-retCh
	return ret.Value, ret.Err
}

func (p *jsClient) eval(expr string) (json.RawMessage, error) {
	return p.send("Vuego.call", h{"name": "eval", "args": []string{expr}}, true)
}

func (p *jsClient) bind(name string, f bindingFunc) error {
	p.Lock()
	_, exists := p.binding[name]
	p.binding[name] = f
	p.Unlock()

	if exists {
		return nil
	}

	if _, err := p.send("Vuego.bind", h{"name": name}, false); err != nil {
		return err
	}
	return nil
}

func (p *jsClient) ready() error {
	if _, err := p.send("Vuego.ready", nil, false); err != nil {
		return err
	}
	return nil
}
