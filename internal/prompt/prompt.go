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
	mu     sync.Mutex
	reader *bufio.Reader
)

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

func Ask(question string, defaultYes bool) bool {
	choice := "y/N"
	if defaultYes {
		choice = "Y/n"
	}
	fmt.Printf("%s [%s]: ", question, choice)

	mu.Lock()
	defer mu.Unlock()
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
