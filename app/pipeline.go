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

func executePipeline(commands []string) error {
	input1, input2 := splitPipeline(commands)

	readEnd, writeEnd, err := os.Pipe()
	if err != nil {
		return err
	}

	cmd1, err := lookupCommand(input1[0], input1[1:])
	if err != nil {
		return err
	}

	cmd2, err := lookupCommand(input2[0], input2[1:])
	if err != nil {
		return err
	}

	// start the pipe first command writes to it, second command reads from it
	cmd1.Stdout = writeEnd
	cmd1.Stderr = os.Stderr
	cmd2.Stdin = readEnd
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr

	// Start both commands running at the same time
	cmd1.Start()
	cmd2.Start()

	// Close our copies of the pipe ends so the commands know when to stop
	// (without this, cmd2 would wait forever for more input)
	writeEnd.Close()
	readEnd.Close()

	// Wait for both commands to finish their work
	cmd1.Wait()
	cmd2.Wait()

	return nil
}
