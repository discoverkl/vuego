package vuego

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
)

type NativeWindow interface {
	Open(url string) error
	Close()
}

type browserNativeWindow struct{}

func (*browserNativeWindow) Open(url string) error {
	openBrowser(url)
	return nil
}

func (*browserNativeWindow) Close() {}

type browserPage struct {
	server    *FileServer
	closeOnce sync.Once
	done      chan struct{}
	win       NativeWindow
}

func NewPage(root http.FileSystem) (Window, error) {
	return NewNativeWindow(root, &browserNativeWindow{}, nil)
}

func NewPageMapURL(root http.FileSystem, mapURL func(net.Listener) string) (Window, error) {
	return NewNativeWindow(root, &browserNativeWindow{}, mapURL)
}

func NewNativeWindow(root http.FileSystem, win NativeWindow, mapURL func(net.Listener) string) (Window, error) {
	// ** local server
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}
	addr := listener.Addr().(*net.TCPAddr)
	log.Println("using port:", addr.Port)

	server := NewFileServer(root)
	server.Listener = listener
	go server.ListenAndServe()

	var url string
	url = fmt.Sprintf("http://localhost:%d", addr.Port)
	if mapURL != nil {
		url = mapURL(listener)
	}

	// ** brower page
	if err := win.Open(url); err != nil {
		return nil, err
	}

	c := &browserPage{
		server: server,
		done:   make(chan struct{}),
		win:    win,
	}

	// 1/2 server.Done() => done
	// 2/2 user call Close() => done
	go func() {
		<-server.Done()
		c.Close()
	}()
	return c, nil
}

// func (c *browserPage) Bind(name string, f interface{}) error {
// 	return c.server.Bind(name, f)
// }

func (c *browserPage) Bind(b Bindings) error {
	return c.server.Bind(b)
}

func (c *browserPage) Eval(js string) Value {
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
