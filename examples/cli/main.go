//go:generate go run github.com/markbates/pkger/cmd/pkger -o cli
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/discoverkl/vuego"
	"github.com/markbates/pkger"
)

var dev = os.Getenv("dev") == "1"

func main() {
	if matchSafeBash() || matchSafeMain() {
		return
	}
	var interactive bool
	var port int
	var tls bool
	var authTable string // csv file with format: username,password
	var redirectPort int
	var workingDir string
	var uid int
	var gid int

	flag.BoolVar(&interactive, "i", false, "run interactive process such as: bash, python")
	flag.IntVar(&port, "p", -1, "binding port")
	flag.IntVar(&uid, "uid", 0, "override user id")
	flag.IntVar(&gid, "gid", 0, "override group id")
	flag.IntVar(&redirectPort, "redirect", 0, "redirect http request on this port")
	flag.BoolVar(&tls, "tls", false, "run https, use 'server.tls' and 'server.key' in working directory")
	flag.StringVar(&authTable, "auth", "", "basic auth csv file with format: username, password")
	flag.StringVar(&workingDir, "w", "", "working directory")
	flag.Parse()
	if port == -1 {
		if tls {
			port = 443
		} else {
			port = 80
		}
	}

	if flag.NArg() == 0 {
		log.Fatal("missing argument")
	}

	name := processName(flag.Args())
	args := processArgs(flag.Args(), interactive)

	vuego.BindFactory("", func(done <-chan bool) interface{} {
		p := &Proc{Name: name, Args: args, WorkingDir: workingDir, Uid: uid, Gid: gid}
		go func() {
			<-done
			p.Close()
		}()

		go func() {
			err := p.run()
			if err != nil {
				log.Println(err)
				p.Close()
			}
		}()

		var vim *Vim
		if len(args) > 0 && args[0] == "bash" {
			vim = &Vim{proc: p}
		}

		return map[string]interface{}{
			"name":   p.name,
			"write":  p.write,
			"listen": p.listen,
			"kill":   p.kill,
			"pwd":    p.pwd,
			"load":   vim.load,
			"save":   vim.save,
		}
	})

	if authTable != "" {
		auth := map[string]string{}
		raw, err := ioutil.ReadFile(authTable)
		if err != nil {
			log.Fatalf("loading auth table failed: %v", err)
		}
		for _, line := range strings.Split(string(raw), "\n") {
			sp := strings.Split(line, ",")
			if len(sp) != 2 {
				continue
			}
			user := strings.TrimSpace(sp[0])
			pass := strings.TrimSpace(sp[1])
			if user == "" {
				continue
			}
			auth[user] = pass
		}
		vuego.Auth = vuego.BasicAuth(func(user, pass string) bool {
			want, ok := auth[user]
			if !ok {
				return false
			}
			return want == pass
		})
	}

	addr := fmt.Sprintf(":%d", port)
	log.Printf("listen on: %s", addr)

	var err error
	if tls {
		if redirectPort != 0 {
			go enforceTLS(redirectPort, port)
		}
		err = vuego.ListenAndServeTLS(addr, pkger.Dir("/cli/fe/dist"), "server.crt", "server.key")
	} else {
		err = vuego.ListenAndServe(addr, pkger.Dir("/cli/fe/dist"))
	}
	if err != nil {
		log.Fatal(err)
	}
}

func processName(args []string) string {
	ret := []string{}
	for i, arg := range args {
		if i == 0 {
			ret = append(ret, filepath.Base(arg))
		} else {
			if arg != "" && arg[0] == '-' {
				continue
			}
			ret = append(ret, arg)
		}
	}
	return strings.Join(ret, " ")
}

func processArgs(args []string, interactive bool) []string {
	ret := []string{}
	if !interactive {
		path, _ := os.Executable()
		ret = append(ret, path, "safe-bash")
	}
	return append(ret, args...)
}

func matchSafeBash() bool {
	if len(os.Args) < 2 || os.Args[1] != "safe-bash" {
		return false
	}

	proc := &SafeBash{Args: os.Args[2:]}
	err := proc.Run()
	if err != nil {
		log.Fatal(err)
	}

	return true
}

func matchSafeMain() bool {
	if len(os.Args) < 2 || os.Args[1] != "safe-main" {
		return false
	}
	if err := safeMain(); err != nil {
		fmt.Println(err) // print to stdout
		os.Exit(-1)
	}
	return true
}

func enforceTLS(httpPort int, tlsPort int) {
	s := http.Server{
		Addr: fmt.Sprintf(":%d", httpPort),
	}
	s.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("https://%s:%d%s", strings.Split(r.Host, ":")[0], tlsPort, r.RequestURI)
		http.Redirect(w, r, url, http.StatusMovedPermanently)
	})
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
