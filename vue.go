package vuego

import (
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

const PageServerPath = "/vuego"

func Bind(name string, f interface{}) error {
	return ins.Bind(name, f)
}

type server struct {
	binding map[string]interface{}
}

func newServer() *server {
	http.Handle(PageServerPath, websocket.Handler(pageServer))
	return &server{
		binding: map[string]interface{}{},
	}
}

func (s *server) Bind(name string, f interface{}) error {
	if err := checkBindFunc(f); err != nil {
		return err
	}
	s.binding[name] = f
	return nil
}

var ins *server

func init() {
	ins = newServer()
}

func pageServer(ws *websocket.Conn) {
	ui, err := newPage(ws)
	if err != nil {
		log.Println("create page error:", err)
	}

	// bind api
	for name, f := range ins.binding {
		err := ui.Bind(name, f)
		if err != nil {
			log.Printf("bind api %s failed: %v", name, err)
		}
	}

	// log.Println(ui.Eval("1 + 2").Int())
	<-ui.Done()
}
