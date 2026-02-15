package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	commands := map[string]interface{}{}
	reader := bufio.NewScanner(os.Stdin)
	fmt.Print("$ ")

	for reader.Scan() {
		text := reader.Text()
		if command, exists := commands[text]; !exists {
			fmt.Println(text + ": command not found")
		} else if exists {
			fmt.Println(command)
		}
		fmt.Print("$ ")

	}
}
