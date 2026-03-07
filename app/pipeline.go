package main

import (
	"os"
)

func hasPipeline(args []string) (hasPipeline bool) {
	for _, arg := range args {
		if arg == "|" {
			return true
		}
	}
	return false
}

func splitPipeline(args []string) (cmd1 []string, cmd2 []string) {
	for i, arg := range args {
		if arg == "|" {
			return args[:i], args[i+1:]
		}
	}
	return args, nil
}

func isBuiltin(cmdName string) bool {
	_, exists := commands[cmdName]
	return exists
}

func executePipeline(commands []string) error {
	input1, input2 := splitPipeline(commands)

	readEnd, writeEnd, err := os.Pipe()
	if err != nil {
		return err
	}

	cmd1Name := input1[0]
	cmd1Args := input1[1:]
	cmd2Name := input2[0]
	cmd2Args := input2[1:]

	cmd1IsBuiltin := isBuiltin(cmd1Name)
	cmd2IsBuiltin := isBuiltin(cmd2Name)

	// Start first command (either builtin or external)
	if cmd1IsBuiltin {
		// Run builtin with stdout redirected to pipe
		originalStdout := os.Stdout
		os.Stdout = writeEnd
		runCommand(cmd1Name, cmd1Args)
		os.Stdout = originalStdout
		writeEnd.Close()
	} else {
		// Start external command
		cmd1, err := lookupCommand(cmd1Name, cmd1Args)
		if err != nil {
			return err
		}
		cmd1.Stdout = writeEnd
		cmd1.Stderr = os.Stderr
		cmd1.Start()
		writeEnd.Close()
		defer cmd1.Wait()
	}

	// Start second command (either builtin or external)
	if cmd2IsBuiltin {
		// Run builtin with stdin redirected from pipe
		originalStdin := os.Stdin
		os.Stdin = readEnd
		runCommand(cmd2Name, cmd2Args)
		os.Stdin = originalStdin
		readEnd.Close()
	} else {
		// Start external command
		cmd2, err := lookupCommand(cmd2Name, cmd2Args)
		if err != nil {
			return err
		}
		cmd2.Stdin = readEnd
		cmd2.Stdout = os.Stdout
		cmd2.Stderr = os.Stderr
		cmd2.Start()
		readEnd.Close()
		cmd2.Wait()
	}

	return nil
}
