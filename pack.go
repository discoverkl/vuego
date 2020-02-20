package vuego

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/here"
	"github.com/markbates/pkger/pkging/embed"
	"github.com/markbates/pkger/pkging/mem"
)

func modNameFromOutput(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	// name from existing package
	dir := filepath.Dir(path)
	proc := exec.Command("go", "list", "-f", "{{.Name}}")
	raw, err := proc.Output()
	if err == nil {
		name := strings.TrimSpace(string(raw))
		if name != "" {
			return name, nil
		}
	}

	// name from parent dir
	pkg := filepath.Base(dir)
	if pkg == "" {
		return "", fmt.Errorf("need package name")
	}
	return pkg, nil
}

// Pack a file or directory to output path, using pkg as pacakge name.
func Pack(path string, output string, pkg string) error {
	if output == "" {
		output = "pkged.go"
	}
	output, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	if f, err := os.Stat(output); err == nil && f.IsDir() {
		return fmt.Errorf("output already exists and is a directory: %s", output)
	}
	if pkg == "" {
		pkg, err = modNameFromOutput(output)
		if err != nil {
			return err
		}
	}
	fmt.Printf("Package: %s\n", pkg)
	fmt.Printf("Output: %s\n", output)

	// ** create memory file system
	info, err := here.Current()
	if err != nil {
		return err
	}
	fs, err := mem.New(info)
	if err != nil {
		return err
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return fs.MkdirAll(path, 0755)
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		return fs.Add(f)
	})
	if err != nil {
		return err
	}

	// ** dump to go source file
	fp, err := os.Create(output)
	if err != nil {
		return err
	}
	defer fp.Close()

	_, err = fmt.Fprintf(fp, `package %s

import (
	"github.com/markbates/pkger"
	"github.com/markbates/pkger/pkging/mem"
)

var _ = pkger.Apply(mem.UnmarshalEmbed([]byte(%s`, pkg, "`")
	if err != nil {
		return err
	}

	b, err := fs.MarshalJSON()
	if err != nil {
		return err
	}

	b, err = embed.Encode(b)
	if err != nil {
		return err
	}

	_, err = fp.Write(b)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(fp, "`)))\n")
	if err != nil {
		return err
	}

	// ** print files
	var mod string
	_ = fs.Walk("", func(path string, info os.FileInfo, err error) error {
		index := strings.Index(path, ":")
		if index != -1 {
			if mod == "" {
				mod = path[0:index]
				fmt.Printf("Module: %s\n", mod)
				fmt.Println("Files:")
			}
			path = path[index+1:]
		}

		if info.IsDir() {
			return nil
		}

		fmt.Println(path)
		return nil
	})

	return nil
}
