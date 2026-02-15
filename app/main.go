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
		"exit": exit,
		"echo": echo,
		"type": typeFunc,
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

func exit(args string) {
	os.Exit(0)
}

func echo(args string) {
	strings.TrimPrefix(args, " ")
	fmt.Println(args)
}

func typeFunc(args string) {
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

func runExternal(name string, args []string) {
	path, err := exec.LookPath(name)
	if err != nil {
		fmt.Println(name + ": command not found")
		return
	}

	cmd := exec.Command(path)
	cmd.Args = append(cmd.Args, args...)
	cmd.Path = path

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Println("error:", err)
	}
}
