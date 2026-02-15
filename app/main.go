package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	commands := map[string]interface{}{
		"exit": exit,
	}
	reader := bufio.NewScanner(os.Stdin)
	fmt.Print("$ ")

	for reader.Scan() {
		text := reader.Text()
		if command, exists := commands[text]; !exists {
			fmt.Println(text + ": command not found")
		} else if exists {
			command.(func())()
		}
		fmt.Print("$ ")

	}
}

func exit() {
	os.Exit(0)
}
