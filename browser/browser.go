package browser

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sync"

	"github.com/discoverkl/vuego"
	"github.com/discoverkl/vuego/one"
)

var dev = one.InDevMode("browser")

type NativeWindow interface {
	Open(url string) error
	Close()
}

type browserNativeWindow struct {}

func (*browserNativeWindow) Open(url string) error {
	openBrowser(url)
	return nil
}

func (*browserNativeWindow) Close(){}

type browserPage struct {
	server    *vuego.FileServer
	closeOnce sync.Once
	done      chan struct{}
	win NativeWindow
}

func NewPage(root http.FileSystem) (vuego.Window, error) {
	return NewNativeWindow(root, &browserNativeWindow{})
}

func NewNativeWindow(root http.FileSystem, win NativeWindow) (vuego.Window, error) {
	// ** local server
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}
	addr := listener.Addr().(*net.TCPAddr)
	url := fmt.Sprintf("http://localhost:%d", addr.Port)
	log.Println("using port:", addr.Port)

	server := vuego.NewFileServer(root)
	server.Listener = listener
	go server.ListenAndServe()

	// ** brower page
	if err := win.Open(url); err != nil {
		return nil, err
	}

	c := &browserPage{
		server: server,
		done: make(chan struct{}),
		win: win,
	}

	// 1/2 server.Done() => done
	// 2/2 user call Close() => done
	go func() {
		<-server.Done()
		c.Close()		
	}()
	return c, nil
}

func (c *browserPage) Bind(name string, f interface{}) error {
	return c.server.Bind(name, f)
}

func (c *browserPage) Eval(js string) vuego.Value {
	panic("Not Implemented")
}

func (c *browserPage) Done() <-chan struct{} {
	return c.done
}

func (c *browserPage) Close() error {
	c.closeOnce.Do(func() {
		if dev {
			log.Println("Window.Close called")
		}
		c.server.Close()
		<-c.server.Done()
		if dev {
			log.Println("Window.server done")
		}

		c.win.Close()

		// notify finally close
		close(c.done)
	})
	return nil
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}
