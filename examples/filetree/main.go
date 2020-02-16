//go:generate go run github.com/markbates/pkger/cmd/pkger -o filetree
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/discoverkl/vuego/ui"
	"github.com/discoverkl/vuego/browser"
	"github.com/discoverkl/vuego/chrome"
	"github.com/markbates/pkger"
)

type Folder struct {
	Name     string    `json:"name"`
	Children []*Folder `json:"children"`
	IsFolder bool      `json:"isFolder"`
}

func openFolder(path string) (*Folder, error) {
	// normalize path
	if path == "" {
		path = "/"
	}
	path = filepath.Clean(path)

	// open folder
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	ret := &Folder{
		Name:     path,
		Children: []*Folder{},
		IsFolder: true,
	}
	for _, f := range files {
		// skip hidden file
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}
		name := filepath.Join(path, f.Name())
		ret.Children = append(ret.Children, &Folder{Name: name, IsFolder: isFolder(f, name)})
	}

	// sort by (type, name)
	sort.Slice(ret.Children, func(i, j int) bool {
		l, r := ret.Children[i], ret.Children[j]
		if l.IsFolder == r.IsFolder {
			return l.Name < r.Name
		}
		return l.IsFolder
	})
	return ret, nil
}

func isFolder(f os.FileInfo, path string) bool {
	if f.Mode()&os.ModeSymlink != 0 {
		target, err := filepath.EvalSymlinks(path)
		if err != nil {
			return false
		}
		targetInfo, err := os.Stat(target)
		if err != nil {
			return false
		}
		return targetInfo.IsDir()
	}
	return f.IsDir()
}

func main() {
	runWebServer()
	// runLocalPage()
	// runNativeApp()
}

var port int

func init() {
	flag.IntVar(&port, "p", 80, "binding port")
	flag.Parse()
}

// run a normal web server
func runWebServer() {
	ui.Bind("openFolder", openFolder)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("listen on: %s", addr)
	if err := ui.ListenAndServe(addr, pkger.Dir("/filetree/fe/dist")); err != nil {
		log.Fatal(err)
	}
}

// run a local web server in background and open its' serving url with your default web browser
func runLocalPage() {
	win, err := browser.NewPage(pkger.Dir("/filetree/fe/dist"))
	if err != nil {
		log.Fatal(err)
	}
	win.Bind("openFolder", openFolder)
	<-win.Done()
}

// run a local web server in background and open its' serving url within a native app (which is a chrome process)
func runNativeApp() {
	win, err := chrome.NewApp(pkger.Dir("/filetree/fe/dist"), 0, 0, 1024, 768)
	if err != nil {
		log.Fatal(err)
	}
	win.Bind("openFolder", openFolder)
	<-win.Done()
}
