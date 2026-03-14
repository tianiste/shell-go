package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

var commandDescriptions = map[string]string{
	"man":     "Display manual pages for commands",
	"help":    "Display information about builtin commands",
	"exit":    "Exit the shell",
	"echo":    "Display a line of text",
	"type":    "Display information about command type",
	"pwd":     "Print the current working directory",
	"cd":      "Change the current directory",
	"history": "Display or manipulate the history list",
}

func initializeCommands() {
	commands = map[string]func(*Command){
		"man":     handleMan,
		"help":    handleHelp,
		"exit":    handleExit,
		"echo":    handleEcho,
		"type":    handleType,
		"pwd":     handlePwd,
		"cd":      handleCd,
		"history": handleHistory,
	}
}

func handleHelp(cmd *Command) {
	if len(cmd.Args) > 0 {
		cmdName := cmd.Args[0]
		if desc, exists := commandDescriptions[cmdName]; exists {
			fmt.Printf("%s: %s\n", cmdName, desc)
		} else {
			fmt.Printf("help: no help topics match '%s'\n", cmdName)
		}
		return
	}

	fmt.Println("Available shell builtins:")

	cmdNames := make([]string, 0, len(commands))
	for cmdName := range commands {
		cmdNames = append(cmdNames, cmdName)
	}
	sort.Strings(cmdNames)

	for _, cmdName := range cmdNames {
		if desc, exists := commandDescriptions[cmdName]; exists {
			fmt.Printf("  %-12s %s\n", cmdName, desc)
		} else {
			fmt.Printf("  %s\n", cmdName)
		}
	}
	fmt.Println("\nType 'help <command>' for more information on a specific command.")
}

func handleExit(cmd *Command) {
	os.Exit(0)
}

func handleEcho(cmd *Command) {
	fmt.Println(strings.Join(cmd.Args, " "))
}

func handleType(cmd *Command) {
	if len(cmd.Args) == 0 {
		return
	}
	command := cmd.Args[0]

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

func handlePwd(cmd *Command) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(dir)
}

func handleCd(cmd *Command) {
	if len(cmd.Args) == 0 {
		return
	}

	path := cmd.Args[0]
	if path == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("cd: %s: No such file or directory\n", path)
			return
		}
		path = homeDir
	}

	if err := os.Chdir(path); err != nil {
		fmt.Printf("cd: %s: No such file or directory\n", path)
	}
}

func createHistoryFile() {
	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		os.WriteFile(historyFile, []byte{}, 0644)
	}
}

func clearHistory() {
	os.WriteFile(historyFile, []byte{}, 0644)
}

func writeToHistory(filePath string) error {
	err := os.WriteFile(filePath, []byte(strings.Join(historyList, "\n")+"\n"), 0644)
	if err != nil {
		return fmt.Errorf("Error writing history file %w", err)
	}
	return nil

}

func appendToHistory(filepath string) error {
	currentFile, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Error opening history file %w", err)
	}
	for _, line := range historyList[lastAppendedIndex:] {
		_, err := currentFile.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("Error while appending history to file %w", err)
		}
	}
	lastAppendedIndex = len(historyList)
	defer currentFile.Close()
	return nil
}

func handleHistory(cmd *Command) {
	if filePath, hasW := cmd.GetFlag("w"); hasW {
		if err := writeToHistory(filePath); err != nil {
			fmt.Println(err)
			return
		}
		return
	}

	if filepath, hasA := cmd.GetFlag("a"); hasA {
		if err := appendToHistory(filepath); err != nil {
			fmt.Println(err)
			return
		}
		return
	}

	if filePath, hasR := cmd.GetFlag("r"); hasR {
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println("Error reading history file:", err)
			return
		}
		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				historyList = append(historyList, line)
			}
		}
		return
	}

	lines := historyList
	startIndex := 0

	var limit int
	var err error

	if nValue, hasN := cmd.GetFlag("n"); hasN {
		limit, err = strconv.Atoi(nValue)
	} else if len(cmd.Args) > 0 {
		limit, err = strconv.Atoi(cmd.Args[0])
	}

	if err == nil && limit > 0 && limit < len(lines) {
		startIndex = len(lines) - limit
		lines = lines[startIndex:]
	}

	for i, line := range lines {
		fmt.Printf("    %d  %s\n", startIndex+i+1, line)
	}
}
