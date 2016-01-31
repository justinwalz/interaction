// Adapted from https://github.com/google/seesaw/blob/master/binaries/seesaw_cli/main.go
package main

import (
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

var (
	command = flag.String("c", "", "Command to execute")

	oldTermState *terminal.State
	prompt       string
	term         *terminal.Terminal
)

func exit() {
	if oldTermState != nil {
		terminal.Restore(syscall.Stdin, oldTermState)
	}
	fmt.Printf("\n")
	os.Exit(0)
}

func fatalf(format string, a ...interface{}) {
	if oldTermState != nil {
		terminal.Restore(syscall.Stdin, oldTermState)
	}
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func suspend() {
	if oldTermState != nil {
		terminal.Restore(syscall.Stdin, oldTermState)
	}
	go resume()
	syscall.Kill(os.Getpid(), syscall.SIGTSTP)
}

func resume() {
	time.Sleep(1 * time.Second)
	fmt.Println("resuming...")
	terminalInit()
}

func terminalInit() {
	var err error
	oldTermState, err = terminal.MakeRaw(syscall.Stdin)
	if err != nil {
		fatalf("Failed to get raw terminal: %v", err)
	}

	term = terminal.NewTerminal(os.Stdin, prompt)
	term.AutoCompleteCallback = autoComplete
}

// autoComplete attempts to complete the user's input when certain
// characters are typed.
func autoComplete(line string, pos int, key rune) (string, int, bool) {
	switch key {
	case 0x01: // Ctrl-A
		return line, 0, true
	case 0x03: // Ctrl-C
		exit()
	case 0x05: // Ctrl-E
		return line, len(line), true
	// case 0x09: // Ctrl-I (Tab)
	// 	_, _, chain, args := cli.FindCommand(string(line))
	// 	line := commandChain(chain, args)
	// 	return line, len(line), true
	case 0x15: // Ctrl-U
		return "", 0, true
	case 0x1a: // Ctrl-Z
		suspend()
		// case '?':
		// 	cmd, subcmds, chain, args := cli.FindCommand(string(line[0:pos]))
		// 	if cmd == nil {
		// 		term.Write([]byte(prompt))
		// 		term.Write([]byte(line))
		// 		term.Write([]byte("?\n"))
		// 	}
		// 	if subcmds != nil {
		// 		for _, c := range *subcmds {
		// 			term.Write([]byte(" " + c.Command))
		// 			term.Write([]byte("\n"))
		// 		}
		// 	} else if cmd == nil {
		// 		term.Write([]byte("Unknown command.\n"))
		// 	}
		//
		// 	line := commandChain(chain, args)
		// 	return line, len(line), true
	}
	return "", 0, false
}

// interactive invokes the interactive CLI interface.
func interactive() {
	// load any configuration, status or otherwise
	// version, err := interaction.Version()
	// if err != nil {
	// 	fatalf("Failed to get version: %v", err)
	// }
	// fmt.Printf("\nJW CLI - version %d\n\n", version)

	u, err := user.Current()
	if err != nil {
		fatalf("Failed to get current user: %v", err)
	}

	prompt = fmt.Sprintf("%s@%s> ", u.Username, "interaction")

	// Setup signal handler before we switch to a raw terminal.
	sigc := make(chan os.Signal, 3)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		<-sigc
		exit()
	}()

	terminalInit()

	for {
		cmdline, err := term.ReadLine()
		if err != nil {
			break
		}
		cmdline = strings.TrimSpace(cmdline)
		if cmdline == "" {
			continue
		}
		if err := Execute(cmdline); err != nil {
			fmt.Println(err)
		}
	}
}

func main() {
	flag.Parse()

	// create context (parse flags, etc)
	// create connections

	if *command == "" {
		interactive()
		exit()
	}

	// if err := seesawCLI.Execute(*command); err != nil {
	// 	fatalf("%v", err)
	// }
	if err := Execute(*command); err != nil {
		fatalf("%v", err)
	}

}

// Execute is a sample command handler
func Execute(command string) error {
	fmt.Println(strings.ToUpper(command))
	return nil
}
