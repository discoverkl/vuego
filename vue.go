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
	"unicode"

	"github.com/discoverkl/vuego/one"
	"golang.org/x/net/websocket"
)

var dev = one.InDevMode("vuego")

// ReadyFuncName is an async ready function in api object.
const ReadyFuncName = "Vuego"
const ContextBindingName = "context"

type ObjectFactory func() interface{}

var Prefix string

// Bind one api for javascript.
func Bind(name string, f interface{}) {
	err := DefaultServer.Bind(name, f)
	if err != nil {
		panic(err)
	}
}

// BindObject bind all public members for javascript.
// If name is empty, bind directly to the api object.
func BindObject(name string, i interface{}) {
	err := DefaultServer.BindObject(name, i)
	if err != nil {
		panic(err)
	}
}

// BindObjectFactory call BindObject for every client session.
func BindObjectFactory(name string, factory ObjectFactory) {
	err := DefaultServer.BindObjectFactory(name, factory)
	if err != nil {
		panic(err)
	}
}

// Done chan is closed when some client had connected and all clients are gone now.
func Done() <-chan struct{} {
	return DefaultServer.Done()
}

func ListenAndServe(addr string, root http.FileSystem) error {
	DefaultServer.Prefix = Prefix
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
	Prefix     string          // path prefix

	server   *http.Server
	serveMux *http.ServeMux

	binding map[string]interface{}
	bindingFactory map[string]ObjectFactory

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
		bindingFactory: map[string]ObjectFactory{},
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
	prefix := s.Prefix
	if prefix != "" && prefix[0] != '/' {
		panic(fmt.Sprintf("Prefix must start with '/', got: %s", prefix))
	}
	prefix = strings.TrimRight(prefix, "/")
	if dev {
		log.Printf("with prefix: %s", prefix)
	}

	s.handleVuego(prefix)

	s.serveMux.Handle(prefix+"/", http.StripPrefix(prefix, http.FileServer(s.root)))
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

func (s *FileServer) handleVuego(prefix string) {
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
		for name, _ := range s.binding {
			names = append(names, name)
		}

		clientScript := injectOptions(&jsOption{
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

func (s *FileServer) Bind(name string, f interface{}) error {
	if err := checkBindFunc(name, f); err != nil {
		return err
	}
	s.binding[name] = f
	return nil
}

func (s *FileServer) BindObjectFactory(name string, factory func() interface{}) error {
	if factory == nil {
		return fmt.Errorf("argument factory is required")
	}
	s.bindingFactory[name] = factory
	return nil
}

type member struct {
	Name  string
	Value reflect.Value
}

func (s *FileServer) BindObject(name string, i interface{}) error {
	binds, err := s.getBindings(name, i)
	if err != nil {
		return err
	}
	for name, f := range binds {
		_ = s.Bind(name, f)
	}
	return nil
}

func (s *FileServer) getBindings(name string, i interface{}) (map[string]interface{}, error) {
	if i == nil {
		return nil, fmt.Errorf("getBindings on nil")
	}
	ret := map[string]interface{}{}
	raw := reflect.ValueOf(i)
	v := raw
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		members := []member{}
		for i := 0; i < v.Type().NumField(); i++ {
			members = append(members, member{
				Name:  v.Type().Field(i).Name,
				Value: v.Field(i),
			})
		}

		for i := 0; i < raw.Type().NumMethod(); i++ {
			members = append(members, member{
				Name:  raw.Type().Method(i).Name,
				Value: raw.Method(i),
			})
		}

		for _, f := range members {
			if !unicode.IsUpper(rune(f.Name[0])) {
				continue
			}
			// convert to js binding name
			fname := fmt.Sprintf("%s%s", strings.ToLower(f.Name[0:1]), f.Name[1:])
			if name != "" {
				fname = fmt.Sprintf("%s.%s", name, fname)
			}

			// wrap none-Func fields and bind methods
			fv := f.Value
			var bindingFunc interface{}
			switch fv.Kind() {
			case reflect.Func:
				bindingFunc = fv.Interface()
			default:
				bindingFunc = func() interface{} { return fv.Interface() }
			}
			ret[fname] = bindingFunc
		}
	default:
		return nil, fmt.Errorf("unsupport object kind: %v", v.Kind())
	}
	return ret, nil
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

	// bind global api
	err = p.BindMap(s.binding)
	if err != nil {
		log.Printf("bind api failed: %v", err)
	}

	// bind session api
	binds := map[string]interface{}{}
	for objName, factory := range s.bindingFactory {
		objBinds, err := s.getBindings(objName, factory())
		if err != nil {
			log.Printf("get session bindings failed: %v", err)
			continue
		}
		for name, f := range objBinds {
			binds[name] = f
		}
	}
	err = p.BindMap(binds)
	if err != nil {
		log.Printf("bind session api failed: %v", err)
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
