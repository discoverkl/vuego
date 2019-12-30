package one

import (
	"fmt"
	"log"
	"net"
	r "reflect"
	"sync/atomic"

	"google.golang.org/grpc"
)

// Server is a grpc server wrapper.
type Server struct {
	port int
	ip   string

	conf *serverConfig
	*grpc.Server
}

type serverConfig struct {
	silent bool
	ip     string
}

// ServerOption provide options.
type ServerOption func(*grpc.Server, *serverConfig) error

// NewServer create a Server instance.
func NewServer(port int, options ...ServerOption) (*Server, error) {
	conf := &serverConfig{}
	rpcServer := grpc.NewServer()

	// register
	for _, option := range options {
		err := option(rpcServer, conf)
		if err != nil {
			return nil, err
		}
	}

	var ip string
	if conf.ip != "" {
		ip = conf.ip
	} else {
		ip = IP()
		if ip == "" {
			log.Println("detect local IP failed, using 0.0.0.0")
			ip = "0.0.0.0"
		}
	}

	s := &Server{
		ip:     ip,
		port:   port,
		conf:   conf,
		Server: rpcServer,
	}
	return s, nil
}

// RunServer create and run a Server.
func RunServer(port int, options ...ServerOption) error {
	s, err := NewServer(port, options...)
	if err != nil {
		return err
	}
	return s.Run()
}

// Run server.
func (s *Server) Run() error {
	// listener
	if !s.conf.silent {
		log.Printf("listen %s", s.Addr())
	}

	lis, err := net.Listen("tcp", s.Addr())
	if err != nil {
		return err
	}
	return s.Serve(lis)
}

// Addr listening.
func (s *Server) Addr() string {
	return fmt.Sprintf("%s:%d", s.ip, s.port)
}

// Client is a grpc client wrapper.
type Client struct {
	addr          string
	newClientFunc interface{}

	version int64            // update on refresh
	conn    *grpc.ClientConn // update on refresh
	stub    interface{}      // update on refresh
}

// NewClient create a Client.
func NewClient(addr string, newClientFunc interface{}) (*Client, error) {
	c := &Client{
		addr:          addr,
		newClientFunc: newClientFunc,
		conn:          nil,
		stub:          nil,
	}

	err := c.Refresh()
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Close internal connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Stub returns internal client interface.
func (c *Client) Stub() interface{} {
	return c.stub
}

// API get api wrapper object from stub object.
func (c *Client) API(getFunc interface{}) interface{} {
	vGet := r.ValueOf(getFunc)
	if vGet.Kind() != r.Func {
		panic("getFunc must be a fucntion")
	}
	ret := vGet.Call([]r.Value{r.ValueOf(c)})
	vapi := ret[0]
	return vapi.Interface()
}

// Refresh reconnect to server.
// It is not thread safe.
func (c *Client) Refresh() error {
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			log.Printf("refresh close old connection: %v", err)
		}
	}

	// connection
	conn, err := grpc.Dial(c.addr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	// client
	vNew := r.ValueOf(c.newClientFunc)
	if vNew.Kind() != r.Func {
		return fmt.Errorf("newClientFunc must be a function")
	}
	ret := vNew.Call([]r.Value{r.ValueOf(conn)})
	vStub := ret[0]

	atomic.AddInt64(&c.version, 1)
	c.conn, c.stub = conn, vStub.Interface()
	// log.Printf("version: %v", c.version)
	return nil
}

// Version number.
func (c *Client) Version() int64 {
	return c.version
}

// Register option.
func Register(registerFunc interface{}, ins interface{}) ServerOption {
	return func(s *grpc.Server, conf *serverConfig) error {
		vf := r.ValueOf(registerFunc)
		if vf.Kind() != r.Func {
			return fmt.Errorf("registerFunc must be a function")
		}
		vf.Call([]r.Value{r.ValueOf(s), r.ValueOf(ins)})
		return nil
	}
}

// Slient option.
func Slient() ServerOption {
	return func(s *grpc.Server, conf *serverConfig) error {
		conf.silent = true
		return nil
	}
}

// BindIP option.
func BindIP(ip string) ServerOption {
	return func(s *grpc.Server, conf *serverConfig) error {
		conf.ip = ip
		return nil
	}
}
