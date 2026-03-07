package main

import (
	"fmt"
	"os"
	"os/exec"
)

func lookupCommand(name string, args []string) (*exec.Cmd, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(path, args...)
	cmd.Args[0] = name
	return cmd, nil
}

func runExternal(name string, args []string) {
	cmd, err := lookupCommand(name, args)
	if err != nil {
		fmt.Println(name + ": command not found")
		return
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	cmd.Run()
}
