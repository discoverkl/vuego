package vuego

import (
	"encoding/json"
	"errors"
	"fmt"
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
}

func newJSClient(ws *websocket.Conn) (*jsClient, error) {
	p := &jsClient{
		ws:      ws,
		pending: map[int]chan result{},
		binding: map[string]bindingFunc{},
	}
	go p.readLoop()
	return p, nil
}

func (p *jsClient) readLoop() {
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

			jsString := func(v interface{}) string { b, _ := json.Marshal(v); return string(b) }
			go func() {
				jsRet, jsErr := "", `""`
				if ret, err := binding(call.Args); err != nil {
					jsErr = jsString(err.Error())
				} else if bytesRet, err := json.Marshal(ret); err != nil {
					jsErr = jsString(err.Error())
				} else {
					jsRet = string(bytesRet)
				}
				expr := fmt.Sprintf(`let root = window.vuego;
if (%[4]s) {
	root['%[1]s']['errors'].get(%[2]d)(%[4]s);
} else {
	root['%[1]s']['callbacks'].get(%[2]d)(%[3]s);
}
root['%[1]s']['callbacks'].delete(%[2]d);
root['%[1]s']['errors'].delete(%[2]d)
`, call.Name, call.Seq, jsRet, jsErr)
				_, err = p.send("Vuego.call", h{"name": "eval", "args": []string{expr}}, true)
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
	log.Printf("send method %s, wait=%v", method, wait)
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