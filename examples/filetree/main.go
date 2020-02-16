//go:generate go run github.com/markbates/pkger/cmd/pkger -o filetree
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/discoverkl/vuego/ui"
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
	app := ui.New(
		ui.Root(pkger.Dir("/filetree/fe/dist")),
		ui.OnlinePort(port),
	)

	app.BindFunc("openFolder", openFolder)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

var port int

func init() {
	flag.IntVar(&port, "p", 80, "binding port")
	flag.Parse()
}
