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
		"cd":   handleCd,
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
		args = normaliseString(args)
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

func normaliseString(input string) (output string) {
	return strings.Trim(input, " ")
}

func handleExit(args string) {
	os.Exit(0)
}

func parseCommandLine(line string) ([]string, error) {
	var args []string
	var current strings.Builder
	inSingleQuotes := false
	hasData := false

	flush := func() {
		if hasData {
			args = append(args, current.String())
			current.Reset()
			hasData = false
		}
	}

	for _, r := range line {
		if r == '\'' {
			inSingleQuotes = !inSingleQuotes
			continue
		}

		if !inSingleQuotes && (r == ' ' || r == '\t') {
			flush()
			continue
		}

		current.WriteRune(r)
		hasData = true
	}

	if inSingleQuotes {
		return nil, fmt.Errorf("unclosed single quote")
	}

	flush()
	return args, nil
}

func handleEcho(input string) {
	args, err := parseCommandLine(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	fmt.Println(strings.Join(args, " "))
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

func handleCd(path string) {
	if normaliseString(path) == "~" {
		homeDir, err := os.UserHomeDir() // could be made a global variable so its not re initialised every time a user tries to do cd ~
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
