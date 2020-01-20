package vuego

import (
	"encoding/json"
	"sync"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
	"github.com/discoverkl/vuego/one"
)

var dev = one.InDevMode("vuego")

// Bind a api for javascript.
func Bind(name string, f interface{}) {
	err := ins.Bind(name, f)
	if err != nil {
		panic(err)
	}
}

func FileServer(root http.FileSystem, ops ...option) error {
	conf := &config{
		addr: ":80",
		serverPath: "",
	}
	for _, op := range ops {
		op(conf)
	}
	HandleHTTP(conf.serverPath)
	http.Handle("/", http.FileServer(root))
	return http.ListenAndServe(conf.addr, nil)
}

// Addr option.
func Addr(addr string) option {
	return func (conf *config) {
		conf.addr = addr
	}
}

// ServerPath option.
func ServerPath(path string) option {
	return func (conf *config) {
		conf.serverPath = path
	}
}

func HandleHTTP(serverPath string) {
	once.Do(func() {
		if serverPath == "" {
			serverPath = defaultServerPath
		}
		if serverPath[0] != '/' {
			panic("serverPath must start with '/'")
		}
		http.Handle(serverPath, websocket.Handler(pageServer))
		http.HandleFunc(getScriptPath(serverPath), func(w http.ResponseWriter, req *http.Request) {
			w.Header().Add("Content-Type", "text/javascript")
			jsQuery := fmt.Sprintf("?%s", req.URL.RawQuery)
			bytes, _ := json.Marshal(jsQuery)

			clientScript := mapScript(script, "/vuego", serverPath)
			clientScript = mapScript(clientScript, "let search = undefined", fmt.Sprintf("let search = %s", string(bytes)))
			fmt.Fprint(w, clientScript)
		})
	})
}

// Attach a websocket connection.
func Attach(ws *websocket.Conn) (Page, error) {
	page, err := newPage(ws)
	if err != nil {
		return nil, fmt.Errorf("create page error: %v", err)
	}
	return page, nil
}

var defaultServerPath = "/vuego"
var once sync.Once
var ins *server

type option func(conf *config)

type config struct {
	addr string
	serverPath string
}

type server struct {
	binding map[string]interface{}
}

func newServer() *server {
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
	ins = newServer()
}

func getScriptPath(serverPath string) string {
	return fmt.Sprintf("%s.js", serverPath)
}

func pageServer(ws *websocket.Conn) {
	page, err := Attach(ws)
	if err != nil {
		log.Printf("attach websocket failed: %v", err)
	}

	// bind api
	for name, f := range ins.binding {
		err := page.Bind(name, f)
		if err != nil {
			log.Printf("bind api %s failed: %v", name, err)
		}
	}

	// server ready
	err = page.Ready()
	if err != nil {
		log.Printf("failed to make page ready: %v", err)
	}

	// wait
	<-page.Done()
}
