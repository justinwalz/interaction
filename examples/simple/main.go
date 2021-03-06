package main

import (
	"fmt"
	"strings"

	"github.com/justinwalz/interaction/interaction"
)

func main() {
	i := interaction.Interactive{
		ExecuteFunc: execute,
		// Prompt:      "> ",
	}
	i.Start()
}

func execute(command string) error {
	fmt.Println(strings.ToUpper(command))
	return nil
}
