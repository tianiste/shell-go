package main

import (
	"os"
	"os/exec"
)

func hasPipeline(args []string) (hasPipeline bool) {
	for _, arg := range args {
		if arg == "|" {
			return true
		}
	}
	return false
}

func splitPipeline(args []string) [][]string {
	var commands [][]string
	var current []string

	for _, arg := range args {
		if arg == "|" {
			if len(current) > 0 {
				commands = append(commands, current)
				current = []string{}
			}
		} else {
			current = append(current, arg)
		}
	}

	if len(current) > 0 {
		commands = append(commands, current)
	}

	return commands
}

func isBuiltin(cmdName string) bool {
	_, exists := commands[cmdName]
	return exists
}

func executePipeline(args []string) error {
	segments := splitPipeline(args)
	n := len(segments)

	if n == 0 {
		return nil
	}

	pipes := make([]*os.File, (n-1)*2)
	for i := 0; i < n-1; i++ {
		readEnd, writeEnd, err := os.Pipe()
		if err != nil {
			return err
		}
		pipes[i*2] = readEnd
		pipes[i*2+1] = writeEnd
	}

	var externalCmds []*exec.Cmd

	for i := 0; i < n; i++ {
		cmdName := segments[i][0]
		cmdArgs := segments[i][1:]

		var input, output *os.File

		if i == 0 {
			input = os.Stdin
		} else {
			input = pipes[(i-1)*2]
		}

		if i == n-1 {
			output = os.Stdout
		} else {
			output = pipes[i*2+1]
		}

		if isBuiltin(cmdName) {
			originalStdin := os.Stdin
			originalStdout := os.Stdout

			os.Stdin = input
			os.Stdout = output

			runCommand(cmdName, cmdArgs)

			os.Stdin = originalStdin
			os.Stdout = originalStdout

			if input != os.Stdin {
				input.Close()
			}
			if output != os.Stdout {
				output.Close()
			}
		} else {
			cmd, err := lookupCommand(cmdName, cmdArgs)
			if err != nil {
				return err
			}

			cmd.Stdin = input
			cmd.Stdout = output
			cmd.Stderr = os.Stderr

			if err := cmd.Start(); err != nil {
				return err
			}

			externalCmds = append(externalCmds, cmd)

			if input != os.Stdin {
				input.Close()
			}
			if output != os.Stdout {
				output.Close()
			}
		}
	}

	for _, cmd := range externalCmds {
		cmd.Wait()
	}

	return nil
}
