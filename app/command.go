package main

import "strings"

type Command struct {
	Name  string
	Flags map[string]string
	Args  []string
}

func ParseCommand(parts []string) *Command {
	if len(parts) == 0 {
		return nil
	}

	cmd := &Command{
		Name:  parts[0],
		Flags: make(map[string]string),
		Args:  []string{},
	}

	i := 1
	for i < len(parts) {
		arg := parts[i]

		if strings.HasPrefix(arg, "-") {
			flagName := strings.TrimPrefix(arg, "-")

			if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "-") {
				cmd.Flags[flagName] = parts[i+1]
				i += 2
			} else {
				cmd.Flags[flagName] = ""
				i++
			}
		} else {
			cmd.Args = append(cmd.Args, arg)
			i++
		}
	}

	return cmd
}

func (c *Command) HasFlag(name string) bool {
	_, exists := c.Flags[name]
	return exists
}

func (c *Command) GetFlag(name string) (string, bool) {
	val, exists := c.Flags[name]
	return val, exists
}
