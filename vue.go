package vuego

import (
	"fmt"
	"log"
	"net/http"
	"golang.org/x/net/websocket"
)

// socket server path.
const (
	ServerPath = "/vuego"
)

// ScriptPath is the script serving path.
var ScriptPath string
var ins *server

// Bind a api for javascript.
func Bind(name string, f interface{}) error {
	return ins.Bind(name, f)
}

// Attach a websocket connection.
func Attach(ws *websocket.Conn) (Page, error) {
	page, err := newPage(ws)
	if err != nil {
		return nil, fmt.Errorf("create page error: %v", err)
	}

	// bind api
	for name, f := range ins.binding {
		err := page.Bind(name, f)
		if err != nil {
			return nil, fmt.Errorf("bind api %s failed: %v", name, err)
		}
	}

	// server ready
	err = page.Ready()
	if err != nil {
		return nil, fmt.Errorf("failed to make page ready: %v", err)
	}
	return page, nil
}

type server struct {
	binding map[string]interface{}
}

func newServer() *server {
	http.Handle(ServerPath, websocket.Handler(pageServer))
	http.HandleFunc(ScriptPath, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/javascript")
		fmt.Fprint(w, script)
	})
	// TODO: handle script path
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

func init() {
	ScriptPath = fmt.Sprintf("%s.js", ServerPath)
	ins = newServer()
}

func pageServer(ws *websocket.Conn) {
	page, err := Attach(ws)
	if err != nil {
		log.Println(err)
	}
	<-page.Done()
}
