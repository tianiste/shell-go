package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

var commands map[string]func([]string)

type DoubleTabCompleter struct {
	inner readline.AutoCompleter
	armed bool
}

func (c *DoubleTabCompleter) Do(line []rune, pos int) ([][]rune, int) {
	candidates, offset := c.inner.Do(line, pos)

	if len(candidates) == 0 {
		c.armed = false
		fmt.Print("\a")
		return nil, 0
	}

	if len(candidates) == 1 {
		c.armed = false
		return candidates, offset
	}

	if !c.armed {
		c.armed = true
		fmt.Print("\a")
		return nil, 0
	}

	c.armed = false

	prefix := string(line[pos-offset : pos])
	var names []string
	for _, cand := range candidates {
		fullWord := prefix + string(cand)
		names = append(names, fullWord)
	}
	sort.Strings(names)
	fmt.Println()
	fmt.Println(strings.Join(names, "  "))
	fmt.Printf("$ %s", string(line))

	return nil, 0
}

func main() {
	commands = map[string]func([]string){
		"exit": handleExit,
		"echo": handleEcho,
		"type": handleType,
		"pwd":  handlePwd,
		"cd":   handleCd,
	}

	seenCommands := make(map[string]bool)
	completers := []readline.PrefixCompleterInterface{}

	for command := range commands {
		seenCommands[command] = true
		completers = append(completers, readline.PcItem(command))
	}

	paths := filepath.SplitList(os.Getenv("PATH"))
	for _, path := range paths {
		files, _ := os.ReadDir(path)

		for _, file := range files {
			info, _ := file.Info()
			if !info.IsDir() && info.Mode().Perm()&0111 != 0 {
				name := info.Name()
				if !seenCommands[name] {
					seenCommands[name] = true
					completers = append(completers, readline.PcItem(name))
				}
			}

		}
	}
	base := readline.NewPrefixCompleter(completers...)
	doubleTabCompleter := &DoubleTabCompleter{inner: base}

	reader, err := readline.NewEx(&readline.Config{
		Prompt:       "$ ",
		HistoryFile:  "/tmp/my_shell_history",
		AutoComplete: doubleTabCompleter,
		FuncFilterInputRune: func(r rune) (rune, bool) {
			if r != readline.CharTab {
				doubleTabCompleter.armed = false
			}
			return r, true
		},
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	for {
		text, err := reader.Readline()
		if err == readline.ErrInterrupt {
			continue
		}
		if err == io.EOF {
			break
		}
		parts, err := parseCommandLine(text)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			firstPrint()
			continue
		}
		if strings.Trim(text, " ") == "" {
			firstPrint()
			continue
		}
		if len(parts) == 0 {
			firstPrint()
			continue
		}
		cmd := parts[0]
		args := parts[1:]

		redirectPosition, redirectType, hasRedirect := checkForRedirect(parts)

		if !hasRedirect {
			if builtin, exists := commands[cmd]; exists {
				builtin(args)
			} else {
				runExternal(cmd, args)
			}
			firstPrint()
			continue
		}

		filename := parts[redirectPosition+1]
		args = parts[1:redirectPosition]
		executeWithRedirect(cmd, args, filename, redirectType)
	}
}

func firstPrint() {
	fmt.Print("$ ")
}

func handleExit(args []string) {
	os.Exit(0)
}

func executeWithRedirect(cmd string, args []string, filename string, redirectType string) {
	var (
		file *os.File
		err  error
	)

	switch redirectType {
	case "stdout": // >
		file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	case "stderr": // 2>
		file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	case "appendStdout": // >>
		file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	case "appendStderr": // 2>>
		file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	default:
		fmt.Fprintf(os.Stderr, "unknown redirect type: %s\n", redirectType)
		return
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	switch redirectType {
	case "stdout", "appendStdout":
		originalStdout := os.Stdout
		os.Stdout = file
		defer func() { os.Stdout = originalStdout }()

	case "stderr", "appendStderr":
		originalStderr := os.Stderr
		os.Stderr = file
		defer func() { os.Stderr = originalStderr }()
	}

	if builtin, exists := commands[cmd]; exists {
		builtin(args)
	} else {
		runExternal(cmd, args)
	}
}

func checkForRedirect(args []string) (redirectPosition int, redirectType string, hasRedirect bool) {
	for i, arg := range args {
		if arg == ">" || arg == "1>" {
			return i, "stdout", true
		}
		if arg == "2>" {
			return i, "stderr", true
		}
		if arg == ">>" || arg == "1>>" {
			return i, "appendStdout", true
		}
		if arg == "2>>" {
			return i, "appendStderr", true
		}
	}
	return 0, "", false
}
func parseCommandLine(line string) ([]string, error) {
	var args []string
	var current strings.Builder
	inSingleQuotes := false
	inDoubleQuotes := false
	hasData := false
	escapeNext := false

	flush := func() {
		if hasData {
			args = append(args, current.String())
			current.Reset()
			hasData = false
		}
	}

	for _, r := range line {
		if escapeNext {
			current.WriteRune(r)
			escapeNext = false
			hasData = true
			continue
		}

		if r == '\\' && !inSingleQuotes {
			escapeNext = true
			continue
		}

		if r == '\'' && !inDoubleQuotes {
			inSingleQuotes = !inSingleQuotes
			continue
		}
		if r == '"' && !inSingleQuotes {
			inDoubleQuotes = !inDoubleQuotes
			continue
		}

		if !inSingleQuotes && !inDoubleQuotes && (r == ' ' || r == '\t') {
			flush()
			continue
		}

		current.WriteRune(r)
		hasData = true
	}

	if inSingleQuotes {
		return nil, fmt.Errorf("unclosed single quote")
	}
	if inDoubleQuotes {
		return nil, fmt.Errorf("unclosed double quote")
	}

	flush()
	return args, nil
}

func handleEcho(args []string) {
	fmt.Println(strings.Join(args, " "))
}

func handleType(args []string) {
	if len(args) == 0 {
		return
	}
	command := args[0]
	if _, exists := commands[command]; exists {
		fmt.Println(command, "is a shell builtin")
		return
	}
	if path, err := exec.LookPath(command); err == nil {
		fmt.Println(command, "is", path)
		return
	}
	fmt.Println(command + ": not found")
}

func handlePwd(args []string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(dir)
}

func runExternal(name string, args []string) {
	path, err := exec.LookPath(name)
	if err != nil {
		fmt.Println(name + ": command not found")
		return
	}

	cmd := exec.Command(path, args...)

	cmd.Args[0] = name

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	cmd.Run()
}

func handleCd(args []string) {
	if len(args) == 0 {
		return
	}
	path := args[0]
	if path == "~" {
		homeDir, err := os.UserHomeDir() // could have been made a global variable so its not re initialised every time a user tries to do cd ~
		if err != nil {
			fmt.Printf("cd: %s: No such file or directory \n", path)
			return
		}
		os.Chdir(homeDir)
		return
	}
	if err := os.Chdir(path); err != nil {
		fmt.Printf("cd: %s: No such file or directory \n", path)
		return
	}
}
