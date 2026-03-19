package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

const (
	shellPrompt = "$ "
	//	historyFile   = "/tmp/my_shell_history"
	bellChar      = "\a"
	optionSpacing = "  "
)

var (
	historyFile       = os.Getenv("HISTFILE")
	commands          map[string]func(*Command)
	historyList       []string
	lastAppendedIndex int
)

func initializeHistory() {
	historyFile = os.Getenv("HISTFILE")
	if strings.TrimSpace(historyFile) == "" {
		return
	}

	createHistoryFile()
	if err := appendHistoryFromFile(historyFile); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	lastAppendedIndex = len(historyList)
}

func main() {
	initializeHistory()
	initializeCommands()
	shellCompleter := NewShellCompleter()
	doubleTabCompleter := &DoubleTabCompleter{inner: shellCompleter}

	reader, err := createReadline(doubleTabCompleter)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	runShell(reader)
}

func runShell(reader *readline.Instance) {
	for {
		text, err := reader.Readline()
		if err == readline.ErrInterrupt {
			continue
		}
		if err == io.EOF {
			break
		}

		executeCommand(text)
	}
}

func executeCommand(text string) {
	parts, err := parseCommandLine(text)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		printPrompt()
		return
	}

	if strings.TrimSpace(text) == "" || len(parts) == 0 {
		printPrompt()
		return
	}

	historyList = append(historyList, strings.TrimSpace(text))

	if hasPipeline(parts) {
		if err := executePipeline(parts); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		printPrompt()
		return
	}

	cmd := parts[0]
	args := parts[1:]

	redirectPosition, redirectType, hasRedirect := checkForRedirect(parts)

	if !hasRedirect {
		runCommand(cmd, args)
		printPrompt()
		return
	}

	filename := parts[redirectPosition+1]
	args = parts[1:redirectPosition]
	executeWithRedirect(cmd, args, filename, redirectType)
	printPrompt()
}

func runCommand(cmd string, args []string) {
	if builtin, exists := commands[cmd]; exists {
		parts := append([]string{cmd}, args...)
		commandStruct := ParseCommand(parts)
		builtin(commandStruct)
	} else {
		runExternal(cmd, args)
	}
}

func printPrompt() {
	fmt.Print(shellPrompt)
}
