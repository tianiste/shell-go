package main

import (
	"fmt"
	"sort"
	"strings"
)

type ManualPage struct {
	Name        string
	Synopsis    string
	Description string
	Options     []Option
	Examples    []string
}

type Option struct {
	Flag        string
	Description string
}

var manualPages = map[string]ManualPage{
	"man": {
		Name:     "man",
		Synopsis: "man [COMMAND]",
		Description: `Display manual pages for shell builtin commands.
Without arguments, lists all available commands.
With a command name, displays detailed manual page for that command.`,
		Examples: []string{
			"man           # List all commands",
			"man echo      # Show manual for echo command",
			"man history   # Show manual for history command",
		},
	},
	"help": {
		Name:     "help",
		Synopsis: "help [COMMAND]",
		Description: `Display brief information about builtin commands.
Without arguments, lists all commands with short descriptions.
With a command name, displays a brief description of that command.`,
		Examples: []string{
			"help          # List all commands",
			"help cd       # Show brief help for cd",
		},
	},
	"exit": {
		Name:     "exit",
		Synopsis: "exit",
		Description: `Exit the shell.
Terminates the shell session immediately.`,
		Examples: []string{
			"exit          # Exit the shell",
		},
	},
	"echo": {
		Name:     "echo",
		Synopsis: "echo [STRING...]",
		Description: `Display a line of text.
Writes all arguments to standard output, separated by spaces, followed by a newline.`,
		Examples: []string{
			"echo Hello World        # Output: Hello World",
			"echo \"Hello World\"      # Output: Hello World",
			"echo $HOME              # Output: /home/user",
		},
	},
	"type": {
		Name:     "type",
		Synopsis: "type COMMAND",
		Description: `Display information about command type.
For each command name, indicates whether it is a shell builtin or an external command.
If it's an external command, displays the full path to the executable.`,
		Examples: []string{
			"type cd       # Output: cd is a shell builtin",
			"type ls       # Output: ls is /usr/bin/ls",
			"type invalid  # Output: invalid: not found",
		},
	},
	"pwd": {
		Name:     "pwd",
		Synopsis: "pwd",
		Description: `Print the current working directory.
Displays the absolute pathname of the current working directory.`,
		Examples: []string{
			"pwd           # Output: /home/user/projects",
		},
	},
	"cd": {
		Name:     "cd",
		Synopsis: "cd [DIRECTORY]",
		Description: `Change the current directory.
Changes the current working directory to DIRECTORY.
If no directory is specified, does nothing.
The special directory ~ represents the user's home directory.`,
		Examples: []string{
			"cd /usr/local          # Change to /usr/local",
			"cd ~                   # Change to home directory",
			"cd ..                  # Change to parent directory",
		},
	},
	"history": {
		Name:     "history",
		Synopsis: "history [-n NUM] [-r FILE] [-w FILE] [NUM]",
		Description: `Display or manipulate the history list.
Without options, displays the command history with line numbers.
Can be used to display a limited number of history entries, load history from a file, or write history to a file.`,
		Options: []Option{
			{Flag: "-n NUM", Description: "Display only the last NUM entries"},
			{Flag: "-r FILE", Description: "Read history from FILE and append to current history"},
			{Flag: "-w FILE", Description: "Write current history to FILE"},
			{Flag: "NUM", Description: "Display only the last NUM entries (positional)"},
		},
		Examples: []string{
			"history               # Show all history",
			"history 10            # Show last 10 entries",
			"history -n 5          # Show last 5 entries",
			"history -r file.txt   # Load history from file",
			"history -w file.txt   # Write history to file",
		},
	},
}

func handleMan(cmd *Command) {
	if len(cmd.Args) == 0 {
		listManualPages()
		return
	}

	cmdName := cmd.Args[0]
	if page, exists := manualPages[cmdName]; exists {
		displayManualPage(page)
	} else {
		fmt.Printf("No manual entry for %s\n", cmdName)
	}
}

func listManualPages() {
	fmt.Println("Available manual pages:")
	fmt.Println()

	cmdNames := make([]string, 0, len(manualPages))
	for cmdName := range manualPages {
		cmdNames = append(cmdNames, cmdName)
	}
	sort.Strings(cmdNames)

	for _, cmdName := range cmdNames {
		if page, exists := manualPages[cmdName]; exists {
			// Get first line of description
			firstLine := strings.Split(page.Description, "\n")[0]
			fmt.Printf("  %-12s %s\n", cmdName, firstLine)
		}
	}
	fmt.Println("\nUse 'man <command>' to see detailed manual page.")
}

func displayManualPage(page ManualPage) {
	fmt.Println()
	fmt.Printf("NAME\n")
	fmt.Printf("    %s\n\n", page.Name)

	fmt.Printf("SYNOPSIS\n")
	fmt.Printf("    %s\n\n", page.Synopsis)

	fmt.Printf("DESCRIPTION\n")
	for _, line := range strings.Split(page.Description, "\n") {
		fmt.Printf("    %s\n", line)
	}
	fmt.Println()

	if len(page.Options) > 0 {
		fmt.Printf("OPTIONS\n")
		for _, opt := range page.Options {
			fmt.Printf("    %-15s %s\n", opt.Flag, opt.Description)
		}
		fmt.Println()
	}

	if len(page.Examples) > 0 {
		fmt.Printf("EXAMPLES\n")
		for _, example := range page.Examples {
			fmt.Printf("    %s\n", example)
		}
		fmt.Println()
	}
}
