package vuego

import (
	"encoding/json"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/discoverkl/vuego/one"
	"golang.org/x/net/websocket"
)

var dev = one.InDevMode("vuego")

// ReadyFuncName is an async ready function in api object.
const ReadyFuncName = "Vuego"

// Bind a api for javascript.
func Bind(name string, f interface{}) {
	err := DefaultServer.Bind(name, f)
	if err != nil {
		panic(err)
	}
}

// Done chan is closed when some client had connected and all clients are gone now.
func Done() <-chan struct{} {
	return DefaultServer.Done()
}

func ListenAndServe(addr string, root http.FileSystem) error {
	DefaultServer.root = root
	DefaultServer.Addr = addr
	return DefaultServer.ListenAndServe()
}

var DefaultServer *FileServer
var defaultServerPath = "/vuego"
var once sync.Once

func init() {
	DefaultServer = NewFileServer(nil)
}

type FileServer struct {
	Addr       string
	ServerPath string
	Listener   net.Listener
	root       http.FileSystem // optional for default instance

	server   *http.Server
	serveMux *http.ServeMux

	binding map[string]interface{}

	// local server done
	wg              sync.WaitGroup
	once            sync.Once
	started         chan struct{}
	localServerDone chan struct{}
	doneOnce        sync.Once
}

func NewFileServer(root http.FileSystem) *FileServer {
	serveMux := http.NewServeMux()
	s := &FileServer{
		root:            root,
		serveMux:        serveMux,
		server:          &http.Server{Handler: serveMux},
		binding:         map[string]interface{}{},
		started:         make(chan struct{}),
		localServerDone: make(chan struct{}),
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
		s.closeLocalServer()
	}()
	return s
}

func (s *FileServer) ListenAndServe() error {
	s.handleVuego()

	s.serveMux.Handle("/", http.FileServer(s.root))
	if s.Listener != nil {
		return s.server.Serve(s.Listener)
	}
	addr := s.Addr
	if addr == "" {
		addr = ":80"
	}
	s.server.Addr = addr
	return s.server.ListenAndServe()
}

func (s *FileServer) Shutdown(ctx context.Context) error {
	s.closeLocalServer()
	return s.server.Shutdown(context.Background())
}

func (s *FileServer) Close() error {
	s.closeLocalServer()
	return s.server.Close()
}

func (s *FileServer) closeLocalServer() {
	s.doneOnce.Do(func() {
		close(s.localServerDone)
	})
}

func (s *FileServer) handleVuego() {
	serverPath := s.ServerPath
	if serverPath == "" {
		serverPath = defaultServerPath
	}
	if serverPath[0] != '/' {
		panic("serverPath must start with '/'")
	}

	s.serveMux.Handle(serverPath, websocket.Handler(s.serveClientConn))
	s.serveMux.HandleFunc(getScriptPath(serverPath), func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/javascript")
		jsQuery := fmt.Sprintf("?%s", req.URL.RawQuery)
		bytes, _ := json.Marshal(jsQuery)

		clientScript := mapScript(script, "/vuego", serverPath)
		clientScript = mapScript(clientScript, "let search = undefined", fmt.Sprintf("let search = %s", string(bytes)))
		fmt.Fprint(w, clientScript)
	})
}

func (s *FileServer) Done() <-chan struct{} {
	return s.localServerDone
}

func (s *FileServer) Bind(name string, f interface{}) error {
	if err := checkBindFunc(name, f); err != nil {
		return err
	}
	s.binding[name] = f
	return nil
}

// ready(0) -> started(1+) -> done(0)
func (s *FileServer) serveClientConn(ws *websocket.Conn) {
	s.wg.Add(1)
	defer func() {
		<-time.After(time.Millisecond * 200) // support fast page refresh
		s.wg.Done()
	}()

	s.once.Do(func() {
		close(s.started)
	})

	p, err := newPage(ws)
	if err != nil {
		log.Printf("attach websocket failed: %v", err)
	}

	// bind api
	for name, f := range s.binding {
		err := p.Bind(name, f)
		if err != nil {
			log.Printf("bind api %s failed: %v", name, err)
		}
	}

	// server ready
	err = p.SetReady()
	if err != nil {
		log.Printf("failed to make page ready: %v", err)
	}

	// wait
	<-p.Done()
}

func getScriptPath(serverPath string) string {
	return fmt.Sprintf("%s.js", serverPath)
}
