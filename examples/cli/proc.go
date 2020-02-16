package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/discoverkl/vuego/ui"
)

type WriterFunc func(p []byte) (n int, err error)

func (w WriterFunc) Write(p []byte) (n int, err error) {
	return w(p)
}

// process on linux/darwin system
type Proc struct {
	Name       string
	Args       []string
	WorkingDir string
	Uid        int
	Gid        int

	cancel context.CancelFunc
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
}

func (p *Proc) Close() error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

func (p *Proc) run() error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("windows platform is not supported")
	}
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	proc := exec.CommandContext(ctx, p.Args[0], p.Args[1:]...)
	proc.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if p.Uid != 0 || p.Gid != 0 {
		proc.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(p.Uid), Gid: uint32(p.Gid)}
	}
	proc.Dir = p.WorkingDir
	p.cmd = proc

	stdin, e1 := proc.StdinPipe()
	stdout, e2 := proc.StdoutPipe()
	stderr, e3 := proc.StderrPipe()
	if e1 != nil || e2 != nil || e3 != nil {
		return fmt.Errorf("pipe subprocess failed")
	}
	p.stdin = stdin
	p.stdout = stdout
	p.stderr = stderr

	err := proc.Start()
	if err != nil {
		return err
	}
	if dev {
		log.Printf("pid: %d", proc.Process.Pid)
	}

	return proc.Wait()
}

func (p *Proc) name() string {
	return p.Name
}

func (p *Proc) write(s string) error {
	if p.cmd == nil {
		return fmt.Errorf("invalid operation: subprocess is not ready")
	}
	_, err := fmt.Fprint(p.stdin, s)
	return err
}

func (p *Proc) listen(writer *ui.Function) error {
	if p.cmd == nil {
		return fmt.Errorf("invalid operation: subprocess is not ready")
	}

	if dev {
		log.Println("[enter] listen")
		defer log.Println("[exit] listen")
	}
	var wg sync.WaitGroup
	pipeCh := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := io.Copy(WriterFunc(func(p []byte) (n int, err error) {
			if dev && len(p) <= 20 {
				log.Printf("maybe command: %v", p)
			}
			v := writer.Call(string(p), 1)
			if v.Err() != nil {
				return 0, v.Err()
			}
			return len(p), nil
		}), p.stdout)
		pipeCh <- err
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := io.Copy(WriterFunc(func(p []byte) (n int, err error) {
			if dev && len(p) <= 20 {
				log.Printf("maybe command: %v", p)
			}
			v := writer.Call(string(p), 2)
			if v.Err() != nil {
				return 0, v.Err()
			}
			return len(p), nil
		}), p.stderr)
		pipeCh <- err
	}()

	go func() {
		wg.Wait()
		close(pipeCh)
	}()

	var err error
	for one := range pipeCh {
		if one != nil {
			err = one
		}
	}
	if err != nil {
		return fmt.Errorf("link error: %w", err)
	}
	return nil
}

func (p *Proc) kill(sig syscall.Signal) error {
	if p.cmd == nil {
		return fmt.Errorf("invalid operation: subprocess is not ready")
	}
	ids, err := grepProcessByGroupID(p.cmd.Process.Pid)
	if err != nil {
		return err
	}

	for _, pid := range ids {
		if pid == p.cmd.Process.Pid {
			continue
		}
		_ = syscall.Kill(pid, sig)
	}
	return nil
}

func (p *Proc) pwd() (string, error) {
	if p.cmd == nil {
		return "", fmt.Errorf("invalid operation: subprocess is not ready")
	}
	return grepWorkingDir(p.cmd.Process.Pid)
}

func grepProcessByGroupID(gid int) ([]int, error) {
	proc := exec.Command("pgrep", "-g", fmt.Sprintf("%d", gid))
	raw, err := proc.Output()
	if err != nil {
		return nil, err
	}
	text := strings.TrimSpace(string(raw))
	ret := []int{}
	for _, sid := range strings.Split(text, "\n") {
		id, err := strconv.Atoi(sid)
		if err != nil {
			log.Printf("parse pgrep result failed: %v", err)
			continue
		}
		ret = append(ret, id)
	}
	return ret, nil
}

func grepWorkingDir(pid int) (string, error) {
	proc := exec.Command("lsof", "-p", fmt.Sprintf("%d", pid), "-Fn")
	raw, err := proc.Output()
	if err != nil {
		return "", err
	}
	text := strings.TrimSpace(string(raw))
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line == "fcwd" {
			if i < len(lines) && lines[i+1] != "" {
				return lines[i+1][1:], nil
			}
			break
		}
	}
	return "", nil
}
