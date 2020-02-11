package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"syscall"
)

type Vim struct {
	proc *Proc
}

func (v *Vim) load(path string) (string, error) {
	if v == nil {
		return "", fmt.Errorf("403 forbidden")
	}
	dir, err := v.proc.pwd()
	log.Println("dir:", dir, err)
	if err != nil {
		return "", err
	}
	return runAsUser(v.proc.Uid, v.proc.Gid, dir, []string{"load", path}, "")
}

func (v *Vim) save(path, content string) error {
	if v == nil {
		return fmt.Errorf("403 forbidden")
	}
	dir, err := v.proc.pwd()
	if err != nil {
		return err
	}
	_, err = runAsUser(v.proc.Uid, v.proc.Gid, dir, []string{"save", path}, content)
	return err
}

func runAsUser(uid int, gid int, workingDir string, args []string, input string) (string, error) {
	path, _ := os.Executable()
	args = append([]string{"safe-main"}, args...)
	proc := exec.Command(path, args...)
	proc.Dir = workingDir
	if uid != 0 || gid != 0 {
		proc.SysProcAttr = &syscall.SysProcAttr{}
		proc.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	}
	if input != "" {
		stdin, err := proc.StdinPipe()
		if err != nil {
			return "", err
		}
		go func() {
			fmt.Fprint(stdin, input)
			stdin.Close()
		}()
	}
	raw, err := proc.Output()
	if err != nil {
		if string(raw) != "" {
			return "", errors.New(string(raw))
		}
		return "", err
	}
	return string(raw), nil
}

func safeMain() error {
	args := os.Args[2:]
	action := args[0]

	switch action {
	case "load":
		path := args[1]
		raw, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		fmt.Print(string(raw))
	case "save":
		path := args[1]
		raw, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		// TODO: match old perm
		err = ioutil.WriteFile(path, raw, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
