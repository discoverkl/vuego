package ui

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/shlex"
)

type UI interface {
	Run() error
	// Done() <-chan struct{}
	Bindable
	RunMode
}

type Bindable interface {
	Bind(b Bindings)
	BindPrefix(name string, b Bindings)
	BindFunc(name string, fn interface{})
	BindObject(obj interface{})
	BindMap(m map[string]interface{})
}

type ui struct {
	conf *uiConfig
	runMode
	confError error
	bindings  []Bindings
}

func New(ops ...Option) UI {
	var err error
	var confError error

	conf := defaultUIConfig()
	for _, op := range ops {
		err = op(conf)
		if err != nil && confError == nil {
			confError = fmt.Errorf("ui config: %w", err)
		}
	}

	app := &ui{conf: conf, confError: confError}
	app.useRunMode()
	app.useSpecialEnvSetting()
	return app
}

func (u *ui) Bind(b Bindings) {
	u.bindings = append(u.bindings, b)
}

func (u *ui) BindPrefix(name string, b Bindings) {
	u.Bind(Prefix(name, b))
}

func (u *ui) BindFunc(name string, fn interface{}) {
	u.Bind(Func(name, fn))
}

func (u *ui) BindObject(obj interface{}) {
	u.Bind(Object(obj))
}

func (u *ui) BindMap(m map[string]interface{}) {
	u.Bind(Map(m))
}

func (u *ui) Run() error {
	c := u.conf

	if u.confError != nil {
		return u.confError
	}

	if !c.Quiet {
		log.Println("run mode:", u.runMode)
	}
	if u.runMode.Empty() {
		return fmt.Errorf("run mode is not set")
	}

	var win Window
	var err error

	switch true {
	case u.IsApp():
		if c.AppChromeBinary != "" {
			ChromeBinary = c.AppChromeBinary
		}
		if c.LocalMapURL == nil {
			win = NewApp(c.Root, c.AppX, c.AppY, c.AppWidth, c.AppHeight, c.AppChromeArgs...)
		} else {
			win = NewAppMapURL(c.Root, c.AppX, c.AppY, c.AppWidth, c.AppHeight, c.LocalMapURL, c.AppChromeArgs...)
		}
	case u.IsPage():
		if c.LocalMapURL == nil {
			win = NewPage(c.Root)
		} else {
			win = NewPageMapURL(c.Root, c.LocalMapURL)
		}
	case u.IsOnline():
		svr := NewFileServer(c.Root)
		svr.Addr = c.OnlineAddr
		svr.Listener = c.OnlineListener
		svr.Prefix = c.OnlinePrefix
		svr.Auth = c.OnlineAuth

		for _, b := range u.bindings {
			err = svr.Bind(b)
			if err != nil {
				return err
			}
		}

		if !c.Quiet {
			log.Printf("listen on: %s", svr.Addr)
		}

		if c.OnlineCertFile != "" && c.OnlineKeyFile != "" {
			return svr.ListenAndServeTLS(c.OnlineCertFile, c.OnlineKeyFile)
		}
		return svr.ListenAndServe()
	default:
		return fmt.Errorf("unsupported mode: %v", u)
	}

	for _, b := range u.bindings {
		err = win.Bind(b)
		if err != nil {
			return err
		}
	}

	if c.LocalExitDelay != nil {
		win.SetExitDelay(*c.LocalExitDelay)
	}

	return win.Open()
	// if err = win.Open(); err != nil {
	// 	return err
	// }
	// <-win.Done()
	// return nil
}

func (u *ui) Done() <-chan struct{} {
	panic(nil)
}

//
// private methods
//

func (u *ui) useRunMode() {
	// get mode from env
	mode := os.Getenv("MODE")
	if mode == "" {
		mode = os.Getenv("mode")
	}
	mode = strings.ToLower(mode)
	override := strings.HasSuffix(mode, "!")
	mode = strings.TrimRight(mode, "!")

	if !override && u.conf.Mode != "" {
		mode = u.conf.Mode
	}

	switch mode {
	case "app":
	case "page":
	case "online":
	default:
		mode = "page"
	}
	u.runMode = runMode(mode)
}

func (u *ui) useSpecialEnvSetting() {
	// chrome args
	chromeEnv := os.Getenv("APP_CHROME_ARGS")
	if chromeEnv != "" {
		chromeArgs, err := shlex.Split(chromeEnv)
		if err != nil {
			log.Printf("parse env arguments failed for APP_CHROME_ARGS: %v", err)
		} else {
			u.conf.AppChromeArgs = append(u.conf.AppChromeArgs, chromeArgs...)
		}
	}

	// chrome binary
	chromePathEnv := os.Getenv("APP_CHROME_BINARY")
	if chromePathEnv != "" {
		u.conf.AppChromeBinary = chromePathEnv
	}
}
