package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func initializeCommands() {
	commands = map[string]func([]string){
		"exit":    handleExit,
		"echo":    handleEcho,
		"type":    handleType,
		"pwd":     handlePwd,
		"cd":      handleCd,
		"history": handleHistory,
	}
}

func handleExit(args []string) {
	os.Exit(0)
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

func handleCd(args []string) {
	if len(args) == 0 {
		return
	}

	path := args[0]
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

func writeToHistory(command string) {
	createHistoryFile()
	f, err := os.OpenFile(historyFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error writing to history:", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(command + "\n"); err != nil {
		fmt.Println("Error writing to history:", err)
	}
}

func handleHistory(args []string) {
	createHistoryFile()
	content, err := os.ReadFile(historyFile)
	if err != nil {
		fmt.Println("Error reading history:", err)
		return
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	startIndex := 0

	if len(args) == 1 {
		numberOfLinesToDisplay, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		if numberOfLinesToDisplay < len(lines) {
			startIndex = len(lines) - numberOfLinesToDisplay
			lines = lines[startIndex:]
		}
	}

	for i, line := range lines {
		fmt.Printf("    %d  %s\n", startIndex+i+1, line)
	}
}
