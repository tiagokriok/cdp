package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

var (
	mu        sync.Mutex
	reader    *bufio.Reader
	assumeYes bool
)

func SetAssumeYes(v bool) {
	mu.Lock()
	defer mu.Unlock()
	assumeYes = v
}

func SetInput(r io.Reader) func() {
	mu.Lock()
	defer mu.Unlock()
	prev := reader
	reader = bufio.NewReader(r)
	return func() {
		mu.Lock()
		defer mu.Unlock()
		reader = prev
	}
}

func Confirm(question string, defaultYes bool) bool {
	mu.Lock()
	auto := assumeYes
	mu.Unlock()
	if auto {
		fmt.Printf("%s yes (--yes)\n", strings.TrimLeft(question, "\n"))
		return true
	}
	return Ask(question, defaultYes)
}

func answerWord(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func Ask(question string, defaultYes bool) bool {
	choice := "y/N"
	if defaultYes {
		choice = "Y/n"
	}
	mu.Lock()
	defer mu.Unlock()
	if assumeYes {
		fmt.Printf("%s %s (default, --yes)\n", strings.TrimLeft(question, "\n"), answerWord(defaultYes))
		return defaultYes
	}
	fmt.Printf("%s [%s]: ", question, choice)
	if reader == nil {
		reader = bufio.NewReader(os.Stdin)
	}
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "" {
		return defaultYes
	}
	return line == "y" || line == "yes"
}
