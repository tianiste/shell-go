# Go Mini Shell

A small shell written in Go.

It supports common builtins, external commands, pipes, redirection, history persistence, and tab completion for commands and filenames.

## Features

- Builtins: `help`, `man`, `exit`, `echo`, `type`, `pwd`, `cd`, `history`
- Run external commands from `PATH`
- Pipelines and output redirection
- History loaded from and written to `HISTFILE`
- Tab completion:
  - command names
  - file and directory names

## Run locally

Requirements:

- Go 1.25+

Run directly:

```bash
go run ./app/*.go
```

Build a binary:

```bash
go build -o shell ./app/*.go
./shell
```

## Example

```bash
$ echo hello
hello
$ history
$ ls | wc -l
```

## Project layout

- `app/main.go` - REPL loop and command execution
- `app/parser.go` - command line parsing
- `app/builtins.go` - builtin command handlers
- `app/completion.go` - tab completion logic

