package one

import (
	"fmt"
	"log"
	"reflect"

	"github.com/spf13/cobra"
)

// CobraCommand serve as documentation.
type CobraCommand interface {
}

// Main is used to provide shorter main body.
func Main(cmdlist []CobraCommand, options ...Option) {
	rootCmd := &cobra.Command{}

	for _, option := range options {
		option(rootCmd)
	}

	for _, raw := range cmdlist {
		cmd, ok := raw.(*cobra.Command)
		if !ok {
			log.Fatal(fmt.Errorf("invalid cobra command: %v", raw))
		}
		rootCmd.AddCommand(cmd)
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// Newable is optional.
type Newable interface {
	New()
}

// Runable exit program if Run() returns a non-nil error.
type Runable interface {
	Run() error
}

// Cmd shouble be embed in target command class.
type Cmd struct {
	Args    []string
	Command *cobra.Command
}

// NewCmd create a new command object which has the same type of zero.
func NewCmd(zero interface{}, use string, short string) *cobra.Command {
	T := reflect.TypeOf(zero)
	if T == nil {
		zero = Cmd{}
		T = reflect.TypeOf(zero)
	}
	c := reflect.New(T)

	// **check
	runner, ok := c.Interface().(Runable)
	if !ok {
		panic(fmt.Sprintf("value %v shouble be of type Runable", c.Type()))
	}

	innerCmd := c.Elem().FieldByName("Cmd")
	var pCmd *Cmd
	if !innerCmd.IsValid() {
		innerCmd = c.Elem()
	}
	if innerCmd.Type() != reflect.TypeOf(pCmd).Elem() {
		panic(fmt.Sprintf("value %v.Cmd shouble be of type Cmd", c.Type().Elem()))
	}

	cc := &cobra.Command{
		Use:   use,
		Short: short,
	}
	c.Elem().FieldByName("Command").Set(reflect.ValueOf(cc))

	// **new (optional)
	init, ok := c.Interface().(Newable)
	if ok {
		init.New()
	}

	// **run
	cc.Run = func(inner *cobra.Command, args []string) {
		c.Elem().FieldByName("Args").Set(reflect.ValueOf(args))

		err := runner.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
	return cc
}

func (c *Cmd) Run() error {
	return fmt.Errorf("command %s is not implemented", c.Command.Name())
}

//
// options
//

type Option func(cmd *cobra.Command)

func Desc(value string) Option {
	return func(cmd *cobra.Command) {
		cmd.Short = value
	}
}
