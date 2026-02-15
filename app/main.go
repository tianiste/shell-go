package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var commands map[string]func(string)

func main() {
	commands = map[string]func(string){
		"exit": handleExit,
		"echo": handleEcho,
		"type": handleType,
		"pwd":  handlePwd,
	}
	reader := bufio.NewScanner(os.Stdin)
	firstPrint()

	for reader.Scan() {
		text := reader.Text()
		parts := strings.Fields(text)
		if strings.Trim(text, " ") == "" {
			firstPrint()
			continue
		}
		cmd := parts[0]
		args := text[len(cmd):]
		args = strings.TrimPrefix(args, " ")
		if builtin, exists := commands[cmd]; exists {
			builtin(args)
		} else {
			runExternal(cmd, parts[1:])
		}

		firstPrint()
	}
}

func firstPrint() {
	fmt.Print("$ ")

}

func handleExit(args string) {
	os.Exit(0)
}

func handleEcho(args string) {
	strings.TrimPrefix(args, " ")
	fmt.Println(args)
}

func handleType(args string) {
	if _, exists := commands[args]; exists {
		fmt.Println(args, "is a shell builtin")
		return
	}
	if path, err := exec.LookPath(args); err == nil {
		fmt.Println(args, "is", path)
		return
	}
	fmt.Println(args + ": not found")
}

func handlePwd(args string) {
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

	if err := cmd.Run(); err != nil {
		fmt.Println("error:", err)
	}
}
