package main

import (
	"bufio"
	"fmt"
	"os"
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
		cmd := parts[0]
		if command, exists := commands[cmd]; !exists {
			fmt.Println(cmd + ": command not found")
		} else if exists {
			args := text[len(cmd):]
			args = strings.TrimPrefix(args, " ")
			command(args)
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
	if strings.HasPrefix(args, " ") {
		args = args[1:]
	}
	fmt.Println(args)
}

func typeFunc(args string) {
	if _, exists := commands[args]; exists {
		fmt.Println(args, "is a shell builtin")
		return
	}
	fmt.Println(args + ": command not found")
}
