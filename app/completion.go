package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

type ShellCompleter struct {
	commandNames []string
}

func NewShellCompleter() *ShellCompleter {
	return &ShellCompleter{commandNames: collectCommandNames()}
}

func (c *ShellCompleter) Do(line []rune, pos int) ([][]rune, int) {
	token, tokenStart, tokenIndex := getCompletionContext(line, pos)
	offset := pos - tokenStart

	if tokenIndex == 0 {
		return completeCommands(c.commandNames, token, offset)
	}

	return completeFileNames(token, offset)
}

type DoubleTabCompleter struct {
	inner readline.AutoCompleter
	armed bool
}

func (c *DoubleTabCompleter) Do(line []rune, pos int) ([][]rune, int) {
	candidates, offset := c.inner.Do(line, pos)

	if len(candidates) == 0 {
		c.armed = false
		fmt.Print(bellChar)
		return nil, 0
	}

	if len(candidates) == 1 {
		c.armed = false
		completion := string(candidates[0])
		if completion != "" && !strings.HasSuffix(completion, "/") && !strings.HasSuffix(completion, " ") {
			completion += " "
			candidates[0] = []rune(completion)
		}
		return candidates, offset
	}

	lcp := findLongestCommonPrefix(candidates)
	if len(lcp) > 0 {
		c.armed = false
		return [][]rune{[]rune(lcp)}, offset
	}

	if !c.armed {
		c.armed = true
		fmt.Print(bellChar)
		return nil, 0
	}

	c.armed = false
	displayCompletionOptions(line, pos, offset, candidates)
	return nil, 0
}

func findLongestCommonPrefix(candidates [][]rune) string {
	if len(candidates) == 0 {
		return ""
	}
	if len(candidates) == 1 {
		return string(candidates[0])
	}

	minLen := len(candidates[0])
	for _, cand := range candidates[1:] {
		if len(cand) < minLen {
			minLen = len(cand)
		}
	}

	for i := 0; i < minLen; i++ {
		char := candidates[0][i]
		for _, cand := range candidates[1:] {
			if cand[i] != char {
				return string(candidates[0][:i])
			}
		}
	}

	return string(candidates[0][:minLen])
}

func displayCompletionOptions(line []rune, pos, offset int, candidates [][]rune) {
	prefix := string(line[pos-offset : pos])

	names := make([]string, 0, len(candidates))
	for _, cand := range candidates {
		fullWord := prefix + string(cand)
		names = append(names, fullWord)
	}
	sort.Strings(names)

	fmt.Println()
	fmt.Println(strings.Join(names, optionSpacing))
	fmt.Printf("%s%s", shellPrompt, string(line))
}

func collectCommandNames() []string {
	seenCommands := make(map[string]bool)
	commandNames := make([]string, 0)

	for command := range commands {
		seenCommands[command] = true
		commandNames = append(commandNames, command)
	}

	paths := filepath.SplitList(os.Getenv("PATH"))
	for _, path := range paths {
		files, _ := os.ReadDir(path)
		for _, file := range files {
			info, _ := file.Info()
			if !info.IsDir() && info.Mode().Perm()&0111 != 0 {
				name := info.Name()
				if !seenCommands[name] {
					seenCommands[name] = true
					commandNames = append(commandNames, name)
				}
			}
		}
	}

	sort.Strings(commandNames)
	return commandNames
}

func getCompletionContext(line []rune, pos int) (token string, tokenStart int, tokenIndex int) {
	inSingleQuotes := false
	inDoubleQuotes := false
	escapeNext := false
	inToken := false
	tokenStart = pos

	for i := 0; i < pos; i++ {
		r := line[i]

		if escapeNext {
			escapeNext = false
			if !inToken {
				inToken = true
				tokenStart = i
			}
			continue
		}

		if r == '\\' && !inSingleQuotes {
			escapeNext = true
			if !inToken {
				inToken = true
				tokenStart = i
			}
			continue
		}

		if r == '\'' && !inDoubleQuotes {
			inSingleQuotes = !inSingleQuotes
			if !inToken {
				inToken = true
				tokenStart = i
			}
			continue
		}

		if r == '"' && !inSingleQuotes {
			inDoubleQuotes = !inDoubleQuotes
			if !inToken {
				inToken = true
				tokenStart = i
			}
			continue
		}

		if !inSingleQuotes && !inDoubleQuotes && (r == ' ' || r == '\t') {
			if inToken {
				tokenIndex++
				inToken = false
			}
			continue
		}

		if !inToken {
			inToken = true
			tokenStart = i
		}
	}

	if !inToken {
		return "", pos, tokenIndex
	}

	return string(line[tokenStart:pos]), tokenStart, tokenIndex
}

func completeCommands(commandNames []string, token string, offset int) ([][]rune, int) {
	candidates := make([][]rune, 0)
	for _, command := range commandNames {
		if strings.HasPrefix(command, token) {
			candidates = append(candidates, []rune(command[len(token):]))
		}
	}

	return candidates, offset
}

func completeFileNames(token string, offset int) ([][]rune, int) {
	dirPart := filepath.Dir(token)
	if token == "" {
		dirPart = "."
	}

	namePrefix := token
	if strings.Contains(token, "/") {
		namePrefix = filepath.Base(token)
	}

	searchDir := dirPart
	if searchDir == "" {
		searchDir = "."
	}

	entries, err := os.ReadDir(searchDir)
	if err != nil {
		return nil, offset
	}

	displayPrefix := ""
	if token != "" && strings.Contains(token, "/") {
		displayPrefix = strings.TrimSuffix(token, namePrefix)
	}

	candidates := make([][]rune, 0)
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, namePrefix) {
			continue
		}

		completion := displayPrefix + name
		if entry.IsDir() {
			completion += "/"
		}

		candidates = append(candidates, []rune(completion[len(token):]))
	}

	sort.Slice(candidates, func(i, j int) bool {
		return string(candidates[i]) < string(candidates[j])
	})

	return candidates, offset
}

func createReadline(completer *DoubleTabCompleter) (*readline.Instance, error) {
	return readline.NewEx(&readline.Config{
		Prompt:       shellPrompt,
		HistoryFile:  os.Getenv("HISTFILE"),
		AutoComplete: completer,
		FuncFilterInputRune: func(r rune) (rune, bool) {
			if r != readline.CharTab {
				completer.armed = false
			}
			return r, true
		},
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
}
