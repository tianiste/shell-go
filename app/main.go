package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	commands := map[string]func(string){
		"exit": exit,
		"echo": echo,
	}
	reader := bufio.NewScanner(os.Stdin)
	firstPrint()

	for reader.Scan() {
		text := reader.Text()
		parts := strings.Fields(text)
		cmd := parts[0]
		if command, exists := commands[cmd]; !exists {
			fmt.Println(text + ": command not found")
		} else if exists {
			command(text[len(cmd)+1:])
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
	fmt.Println(args)
}
