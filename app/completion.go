package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

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

func buildCompleters() []readline.PrefixCompleterInterface {
	seenCommands := make(map[string]bool)
	completers := []readline.PrefixCompleterInterface{}

	for command := range commands {
		seenCommands[command] = true
		completers = append(completers, readline.PcItem(command))
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
					completers = append(completers, readline.PcItem(name))
				}
			}
		}
	}

	return completers
}

func createReadline(completer *DoubleTabCompleter) (*readline.Instance, error) {
	return readline.NewEx(&readline.Config{
		Prompt:       shellPrompt,
		HistoryFile:  historyFile,
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
