package vuego

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/discoverkl/vuego/one"
	"golang.org/x/net/websocket"
)

var dev = one.InDevMode("vuego")

// ReadyFuncName is an async ready function in api object.
const ReadyFuncName = "Vuego"
const ContextBindingName = "context"

type UIContext struct {
	Request *http.Request
	Done    <-chan bool
}

type ObjectFactory func(*UIContext) interface{}

// var Prefix string

// var Auth func(http.HandlerFunc) http.HandlerFunc

// // Bind one api for javascript.
// func Bind(name string, f interface{}) {
// 	err := DefaultServer.Bind(name, f)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// // BindObject bind all public members for javascript.
// // If name is empty, bind directly to the api object.
// func BindObject(name string, i interface{}) {
// 	err := DefaultServer.BindObject(name, i)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// // BindFactory call factory and bind its' return value for each client session.
// func BindFactory(name string, factory ObjectFactory) {
// 	err := DefaultServer.BindFactory(name, factory)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// // Done chan is closed when some client had connected and all clients are gone now.
// func Done() <-chan struct{} {
// 	return DefaultServer.Done()
// }

// func ListenAndServe(addr string, root http.FileSystem) error {
// 	DefaultServer.Prefix = Prefix
// 	DefaultServer.Auth = Auth
// 	DefaultServer.root = root
// 	DefaultServer.Addr = addr
// 	return DefaultServer.ListenAndServe()
// }

// func ListenAndServeTLS(addr string, root http.FileSystem, certFile, keyFile string) error {
// 	DefaultServer.Prefix = Prefix
// 	DefaultServer.Auth = Auth
// 	DefaultServer.root = root
// 	DefaultServer.Addr = addr
// 	return DefaultServer.ListenAndServeTLS(certFile, keyFile)
// }

// var DefaultServer *FileServer
var defaultServerPath = "/vuego"
var once sync.Once

// func init() {
// 	DefaultServer = NewFileServer(nil)
// }

type FileServer struct {
	Addr       string
	ServerPath string
	Listener   net.Listener
	root       http.FileSystem // optional for default instance
	Prefix     string          // path prefix
	Auth       func(http.HandlerFunc) http.HandlerFunc

	server   *http.Server
	serveMux *http.ServeMux

	// binding        map[string]interface{}
	// bindingFactory map[string]ObjectFactory
	bindingNames map[string]bool // for js placeholder
	bindings     []Bindings

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
		root:     root,
		serveMux: serveMux,
		server:   &http.Server{Handler: serveMux},
		// binding:         map[string]interface{}{},
		// bindingFactory:  map[string]ObjectFactory{},
		bindingNames:    map[string]bool{},
		bindings:        []Bindings{},
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
	s.installHandlers(false)
	if s.Listener != nil {
		return s.server.Serve(s.Listener)
	}
	return s.server.ListenAndServe()
}

func (s *FileServer) ListenAndServeTLS(certFile, keyFile string) error {
	s.installHandlers(true)
	if s.Listener != nil {
		return s.server.ServeTLS(s.Listener, certFile, keyFile)
	}
	return s.server.ListenAndServeTLS(certFile, keyFile)
}

func (s *FileServer) installHandlers(tls bool) {
	prefix := s.Prefix
	if prefix != "" && prefix[0] != '/' {
		panic(fmt.Sprintf("Prefix must start with '/', got: %s", prefix))
	}
	prefix = strings.TrimRight(prefix, "/")
	if dev {
		log.Printf("with prefix: %s", prefix)
	}
	s.handleVuego(prefix, tls)
	s.serveMux.Handle(prefix+"/", http.StripPrefix(prefix, http.FileServer(s.root)))

	addr := s.Addr
	if addr == "" {
		addr = ":80"
	}
	s.server.Addr = addr

	if s.Auth != nil {
		s.server.Handler = s.Auth(s.serveMux.ServeHTTP)
	}
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

func (s *FileServer) handleVuego(prefix string, tls bool) {
	serverPath := s.ServerPath
	if serverPath == "" {
		serverPath = defaultServerPath
	}
	if serverPath[0] != '/' {
		panic("serverPath must start with '/'")
	}

	s.serveMux.Handle(prefix+serverPath, http.StripPrefix(prefix, websocket.Handler(s.serveClientConn)))
	s.serveMux.Handle(prefix+getScriptPath(serverPath), http.StripPrefix(prefix, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/javascript")
		jsQuery := fmt.Sprintf("?%s", req.URL.RawQuery)
		// bytes, _ := json.Marshal(jsQuery)

		names := []string{}
		for name, _ := range s.bindingNames {
			names = append(names, name)
		}

		clientScript := injectOptions(&jsOption{
			TLS:      tls,
			Prefix:   prefix,
			Search:   jsQuery,
			Bindings: names,
		})
		// clientScript = mapScript(clientScript, `"/vuego"`, fmt.Sprintf(`"%s"`, prefix + serverPath))
		// clientScript = mapScript(clientScript, "let search = undefined", fmt.Sprintf("let search = %s", string(bytes)))
		fmt.Fprint(w, clientScript)
	})))
}

func (s *FileServer) Done() <-chan struct{} {
	return s.localServerDone
}

func (s *FileServer) Bind(b Bindings) error {
	if b.Error() != nil {
		return b.Error()
	}
	if err := s.collectBindNames(b); err != nil {
		return err
	}
	s.bindings = append(s.bindings, b)
	return nil
}

func (s *FileServer) collectBindNames(b Bindings) error {
	for _, name := range b.Names() {
		s.bindingNames[name] = true
	}
	return nil
}

// func (s *FileServer) Bind(name string, f interface{}) error {
// 	if err := s.collectBindNames(name, f); err != nil {
// 		return err
// 	}
// 	s.binding[name] = f
// 	return nil
// }

// func (s *FileServer) BindFactory(name string, factory ObjectFactory) error {
// 	if factory == nil {
// 		return fmt.Errorf("argument factory is required")
// 	}
// 	// preflight check
// 	done := make(chan bool)
// 	defer close(done)
// 	f := factory(&FactoryContext{Request: nil, Done: done})
// 	if err := s.collectBindNames(name, f); err != nil {
// 		return err
// 	}
// 	s.bindingFactory[name] = factory
// 	return nil
// }

// func (s *FileServer) collectBindNames(name string, f interface{}) error {
// 	// preflight check
// 	hold, err := getBindings(name, f)
// 	if err != nil {
// 		return fmt.Errorf("invalid binding: %w", err)
// 	}
// 	for subName, target := range hold {
// 		if err = checkBindFunc(subName, target); err != nil {
// 			return fmt.Errorf("invalid binding: %w", err)
// 		}
// 		s.bindingNames[subName] = true
// 	}
// 	return nil
// }

type member struct {
	Name  string
	Value reflect.Value
}

// ready(0) -> started(1+) -> done(0)
func (s *FileServer) serveClientConn(ws *websocket.Conn) {
	s.wg.Add(1)
	done := make(chan bool)
	defer func() {
		close(done)
	}()
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

	// apply binding
	binds := map[string]BindingFunc{}
	collect := func(objName string, target interface{}) {
		objBinds, err := getBindings(objName, target)
		if err != nil {
			log.Printf("get session bindings failed: %v", err)
			return
		}
		for name, f := range objBinds {
			binds[name] = f
		}
	}

	// for name, target := range s.binding {
	// 	collect(name, target)
	// }
	// for name, factory := range s.bindingFactory {
	// 	collect(name, factory(&UIContext{Request: ws.Request(), Done: done}))
	// }

	c := &UIContext{Request: ws.Request(), Done: done}
	for _, b := range s.bindings {
		for name, target := range b.Map(c) {
			collect(name, target)
		}
	}

	err = p.bindMap(binds)
	if err != nil {
		log.Printf("binding failed: %v", err)
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

func BasicAuth(auth func(user string, pass string) bool) func(http.HandlerFunc) http.HandlerFunc {
	return func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok || !auth(user, pass) {
				w.Header().Set("www-authenticate", `Basic realm="nothing"`)
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, http.StatusText(http.StatusUnauthorized))
				return
			}
			handler(w, r)
		}
	}
}
