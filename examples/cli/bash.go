package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/google/shlex"
)

type SafeBash struct {
	Args []string // command arguments prefix
	cmd  *exec.Cmd
}

func (b *SafeBash) Run() error {
	if len(b.Args) == 0 {
		return fmt.Errorf("missing argument")
	}

	// by pass SIGINT to subprocess (use SIGKILL to exit safe-bash)
	sigCh := make(chan os.Signal)
	defer close(sigCh)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)
	go func() {
		<-sigCh
	}()

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		if b.cmd != nil {
			// input as stdin
			continue
		}

		// input as arguments
		input := sc.Text()
		userArgs, err := shlex.Split(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		args := append(b.Args[1:], userArgs...)
		fmt.Fprintf(os.Stderr, "+ %s %s\n", b.Args[0], strings.Join(args, " "))
		cmd := exec.Command(b.Args[0], args...)

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		b.cmd = cmd
		err = cmd.Wait()
		b.cmd = nil
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	if sc.Err() != nil && sc.Err() != io.EOF {
		return sc.Err()
	}
	return nil
}
