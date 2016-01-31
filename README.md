# Interaction

A golang library to enhance a cli with REPL-esque command execution. Adapted from Google's [seesaw_cli](https://github.com/google/seesaw/blob/master/binaries/seesaw_cli/main.go)

# Usage

You must provide an `execute` function when creating an `Interactor`, which handles commands as they are parsed.

```
func Execute(command string) error
```

You can optionally provide a prompt, and a slice of signals to listen for and initiate an exit.
```
Prompt string
ExitSignals []os.Signal
```

# Examples

See the examples directory for some example usages

* Simple: passes commands to locally defined Execute function (`strings.ToUpper`).
* [TODO] Package: separates out responsibilities for handling connection, context loading, and execution.
