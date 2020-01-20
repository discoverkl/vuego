package vuego

import (
	"encoding/json"
	"time"
	"net"
	"sync"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
	"github.com/discoverkl/vuego/one"
)

var dev = one.InDevMode("vuego")

// ReadyFuncName is an async ready function in api object.
const ReadyFuncName = "Vuego"

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
	if conf.listener != nil {
		return http.Serve(conf.listener, nil)		
	}
	return http.ListenAndServe(conf.addr, nil)
}

// Done chan is closed when some client had connected and all clients are gone now.
func Done() <-chan struct{} {
	return ins.Done()
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

// Listener option.
func Listener(listener net.Listener) option {
	return func (conf *config) {
		conf.listener = listener
	}
}

//TODO: DO NOT use global http.Handle
func HandleHTTP(serverPath string) {
	once.Do(func() {
		if serverPath == "" {
			serverPath = defaultServerPath
		}
		if serverPath[0] != '/' {
			panic("serverPath must start with '/'")
		}
		http.Handle(serverPath, websocket.Handler(ins.pageServer))
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
	listener net.Listener
}

type server struct {
	wg sync.WaitGroup
	binding map[string]interface{}

	once sync.Once
	started chan struct{}
	done chan struct{}
}

func newServer() *server {
	s := &server{
		binding: map[string]interface{}{},
		started: make(chan struct{}),
		done: make(chan struct{}),
	}
	go func() {
		<-s.started
		if dev {
			log.Println("server active")
		}
		s.wg.Wait()
		if dev {
			log.Println("server done")
		}
		close(s.done)
	}()
	return s
}

func (s *server) Done() <-chan struct{} {
	return s.done
}

func (s *server) Bind(name string, f interface{}) error {
	if err := checkBindFunc(name, f); err != nil {
		return err
	}
	s.binding[name] = f
	return nil
}

func (s *server) pageServer(ws *websocket.Conn) {
	s.wg.Add(1)
	defer func() {
		<- time.After(time.Millisecond * 200)		// support fast page refresh
		s.wg.Done()
	}()
	s.once.Do(func() {
		close(s.started)
	})
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

func init() {
	ins = newServer()
}

func getScriptPath(serverPath string) string {
	return fmt.Sprintf("%s.js", serverPath)
}