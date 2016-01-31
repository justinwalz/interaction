// Package interaction creates an interactive REPL for a context.
// Adapted from https://github.com/google/seesaw/blob/master/binaries/seesaw_cli/main.go
package interaction

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

// Interactive defines an interactive terminal
type Interactive struct {
	// ExecuteFunc to execute commands
	ExecuteFunc func(command string) error
	// Prompt to show user in terminal
	Prompt string
	// ExitSignals are the signals to trap for exiting
	ExitSignals []os.Signal

	oldTermState *terminal.State
	term         *terminal.Terminal
}

var (
	command = flag.String("c", "", "Command to execute")
)

func (i *Interactive) exit() {
	if i.oldTermState != nil {
		terminal.Restore(syscall.Stdin, i.oldTermState)
	}
	fmt.Printf("\n")
	os.Exit(0)
}

func (i *Interactive) fatalf(format string, a ...interface{}) {
	if i.oldTermState != nil {
		terminal.Restore(syscall.Stdin, i.oldTermState)
	}
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func (i *Interactive) suspend() {
	if i.oldTermState != nil {
		terminal.Restore(syscall.Stdin, i.oldTermState)
	}
	go i.resume()
	syscall.Kill(os.Getpid(), syscall.SIGTSTP)
}

func (i *Interactive) resume() {
	time.Sleep(1 * time.Second)
	fmt.Println("resuming...")
	i.terminalInit()
}

func (i *Interactive) terminalInit() {
	var err error
	i.oldTermState, err = terminal.MakeRaw(syscall.Stdin)
	if err != nil {
		i.fatalf("Failed to get raw terminal: %v", err)
	}

	i.term = terminal.NewTerminal(os.Stdin, i.Prompt)
	i.term.AutoCompleteCallback = i.defaultAutoComplete
}

// autoComplete attempts to complete the user's input when certain
// characters are typed.
func (i *Interactive) defaultAutoComplete(line string, pos int, key rune) (string, int, bool) {
	switch key {
	case 0x01: // Ctrl-A
		return line, 0, true
	case 0x03: // Ctrl-C
		i.exit()
	case 0x05: // Ctrl-E
		return line, len(line), true
	case 0x15: // Ctrl-U
		return "", 0, true
	case 0x1a: // Ctrl-Z
		i.suspend()
	}
	return "", 0, false
}

// interactive invokes the interactive CLI interface.
func (i *Interactive) interactive() {
	// version := "0.0.1"
	// fmt.Printf("loading nteraction v%s...\n", version)

	if len(i.Prompt) == 0 {
		u, err := user.Current()
		if err != nil {
			i.fatalf("Failed to get current user: %v", err)
		}
		h, err := os.Hostname()
		if err != nil {
			i.fatalf("Failed to get hostname: %v", err)
		}

		i.Prompt = fmt.Sprintf("%s@%s> ", u.Username, h)
	}

	// Setup signal handler before we switch to a raw terminal.
	if len(i.ExitSignals) == 0 {
		i.ExitSignals = []os.Signal{syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM}
	}

	sigc := make(chan os.Signal, len(i.ExitSignals))
	signal.Notify(sigc, i.ExitSignals...)
	go func() {
		<-sigc
		i.exit()
	}()

	i.terminalInit()

	for {
		cmdline, err := i.term.ReadLine()
		if err != nil {
			break
		}
		cmdline = strings.TrimSpace(cmdline)
		if cmdline == "" {
			continue
		}
		if err := i.ExecuteFunc(cmdline); err != nil {
			fmt.Println(err)
		}
	}
}

// Start the interactive terminal
func (i *Interactive) Start() error {
	if i.ExecuteFunc == nil {
		i.fatalf("%v", errors.New("ExecuteFunc must not be nil"))
	}

	// create context (parse flags, etc)
	// create connections
	flag.Parse()

	if *command == "" {
		i.interactive()
		i.exit()
	}

	if err := i.ExecuteFunc(*command); err != nil {
		i.fatalf("%v", err)
	}

	return nil
}
