package chrome

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"github.com/discoverkl/vuego"
	"github.com/discoverkl/vuego/one"
)

var dev = one.InDevMode("chrome")

type chromePage struct {
	cmd        *exec.Cmd
	chromeDone chan struct{}
	closeOnce  sync.Once
}

func New(root http.FileSystem, dir string, x, y int, width, height int, customArgs ...string) (vuego.Window, error) {
	// ** local server
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}
	addr := listener.Addr().(*net.TCPAddr)
	url := fmt.Sprintf("http://localhost:%d", addr.Port)
	log.Println("using port:", addr.Port)
	go vuego.FileServer(root, vuego.Listener(listener))

	// ** native window
	if dir == "" {
		name, err := ioutil.TempDir("", "vuego-chrome")
		if err != nil {
			return nil, err
		}
		dir = name
	}
	args := append(defaultChromeArgs, fmt.Sprintf("--app=%s", url))
	args = append(args, fmt.Sprintf("--user-data-dir=%s", dir))
	args = append(args, fmt.Sprintf("--window-position=%d,%d", x, y))
	args = append(args, fmt.Sprintf("--window-size=%d,%d", width, height))
	args = append(args, customArgs...)

	c, err := newChromeWithArgs(findChrome(), args...)
	if err != nil {
		return nil, err
	}

	// close window when all clients gone
	go func() {
		<-vuego.Done()
		c.Close()
	}()

	// done = chrome gone
	go func() {
		_ = c.cmd.Wait()
		close(c.chromeDone)
	}()

	return c, nil
}

func newChromeWithArgs(chromeBinary string, args ...string) (*chromePage, error) {
	if chromeBinary == "" {
		return nil, fmt.Errorf("could not find chrome in your system")
	}
	c := &chromePage{
		cmd:        exec.Command(chromeBinary, args...),
		chromeDone: make(chan struct{}),
	}

	if err := c.cmd.Start(); err != nil {
		return nil, err
	}
	log.Println("pid:", c.cmd.Process.Pid)

	return c, nil
}

func (c *chromePage) Bind(name string, f interface{}) error {
	panic("Not Implemented")
}

func (c *chromePage) Eval(js string) vuego.Value {
	panic("Not Implemented")
}

func (c *chromePage) Done() <-chan struct{} {
	return c.chromeDone
}

func (c *chromePage) Close() error {
	c.closeOnce.Do(func() {
		if dev {
			log.Println("chromePage.Close called")
		}
		if state := c.cmd.ProcessState; state == nil || !state.Exited() {
			err := c.cmd.Process.Signal(os.Interrupt) // DO NOT kill -> enable gracefully exit
			if err != nil {
				log.Println("kill chrome process error:", err)
			}
		}
		//TODO: timeout and force kill
		<-c.chromeDone
	})
	return nil
}

//
// tool functions
//

var defaultChromeArgs = []string{
	"--disable-background-networking",
	"--disable-background-timer-throttling",
	"--disable-backgrounding-occluded-windows",
	"--disable-breakpad",
	"--disable-client-side-phishing-detection",
	"--disable-default-apps",
	"--disable-dev-shm-usage",
	"--disable-infobars",
	"--disable-extensions",
	"--disable-features=site-per-process",
	"--disable-hang-monitor",
	"--disable-ipc-flooding-protection",
	"--disable-popup-blocking",
	"--disable-prompt-on-repost",
	"--disable-renderer-backgrounding",
	"--disable-sync",
	"--disable-translate",
	"--metrics-recording-only",
	"--no-first-run",
	"--safebrowsing-disable-auto-update",
	"--enable-automation",
	"--password-store=basic",
	"--use-mock-keychain",
}

func findChrome() string {
	var paths []string
	switch runtime.GOOS {
	case "darwin":
		paths = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
		}
	case "windows":
		paths = []string{
			"C:/Users/" + os.Getenv("USERNAME") + "/AppData/Local/Google/Chrome/Application/chrome.exe",
			"C:/Program Files (x86)/Google/Chrome/Application/chrome.exe",
			"C:/Program Files/Google/Chrome/Application/chrome.exe",
			"C:/Users/" + os.Getenv("USERNAME") + "/AppData/Local/Chromium/Application/chrome.exe",
		}
	default:
		paths = []string{
			"/usr/bin/google-chrome-stable",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
		}
	}

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		return path
	}
	return ""
}
