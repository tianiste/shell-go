package main

import (
	"fmt"
	"os"
)

func checkForRedirect(args []string) (position int, redirectType string, hasRedirect bool) {
	for i, arg := range args {
		switch arg {
		case ">", "1>":
			return i, "stdout", true
		case "2>":
			return i, "stderr", true
		case ">>", "1>>":
			return i, "appendStdout", true
		case "2>>":
			return i, "appendStderr", true
		}
	}
	return 0, "", false
}

func executeWithRedirect(cmd string, args []string, filename string, redirectType string) {
	file, err := openRedirectFile(filename, redirectType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	switch redirectType {
	case "stdout", "appendStdout":
		originalStdout := os.Stdout
		os.Stdout = file
		defer func() { os.Stdout = originalStdout }()
	case "stderr", "appendStderr":
		originalStderr := os.Stderr
		os.Stderr = file
		defer func() { os.Stderr = originalStderr }()
	}

	runCommand(cmd, args)
}

func openRedirectFile(filename string, redirectType string) (*os.File, error) {
	switch redirectType {
	case "stdout", "stderr":
		return os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	case "appendStdout", "appendStderr":
		return os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	default:
		return nil, fmt.Errorf("unknown redirect type: %s", redirectType)
	}
}
