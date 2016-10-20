package main

import (
	"bytes"
	"fmt"
	"io"
	"regexp"

	"strings"
	"text/scanner"

	"github.com/peterh/liner"
)

const (
	promptContinue = "....."
	indent         = "    "
)

type contLiner struct {
	*liner.State
	buffer string
	depth  int
}

func newContLiner() *contLiner {
	rl := liner.NewLiner()
	rl.SetCtrlCAborts(true)
	return &contLiner{State: rl}
}

func (cl *contLiner) promptString(p string) string {
	if cl.buffer != "" {
		return promptContinue + strings.Repeat(indent, cl.depth)
	}

	return p
}

func (cl *contLiner) Prompt(p string) (string, error) {
	line, err := cl.State.Prompt(cl.promptString(p))
	switch err {
	case io.EOF:
		println()
	case liner.ErrPromptAborted:
		if cl.buffer != "" {
			cl.Accepted()
		} else {
			println("(^D to quit)")
		}
		err = nil
	//case liner.ErrPromptCtrlZ:
	//	go exec.Command("kill", "-SIGTSTP", strconv.Itoa(os.Getpid())).Run()
	//	err = nil
	case nil:
		if cl.buffer != "" {
			cl.buffer += "\n" + line
		} else {
			cl.buffer = line
		}
	}

	return cl.buffer, err
}

func (cl *contLiner) Accepted() {
	cl.buffer = regexp.MustCompile(`\n+`).ReplaceAllString(cl.buffer, "\n")

	if p := strings.Index(cl.buffer, "\n"); p != -1 {
		cl.buffer = cl.buffer[:p] + cl.buffer[p+1:]
	}
	if p := strings.LastIndex(cl.buffer, "\n"); p != -1 {
		cl.buffer = cl.buffer[:p] + cl.buffer[p+1:]
	}
	cl.State.AppendHistory(strings.Replace(cl.buffer, "\n", ";", -1))
	cl.buffer = ""
}

func (cl *contLiner) Reindent() {
	oldDepth := cl.depth
	cl.depth = cl.countDepth()

	if cl.depth < oldDepth {
		lines := strings.Split(cl.buffer, "\n")
		if len(lines) > 1 {
			lastLine := lines[len(lines)-1]
			fmt.Print("\x1b[1A") // Cursor up one
			fmt.Printf("\r%s%s", cl.promptString(""), lastLine)
			fmt.Print("\x1b[0K") // Erase to right
			fmt.Print("\n")
		}
	}
}

func (cl *contLiner) countDepth() int {
	depth := 0

	s := new(scanner.Scanner)
	s.Init(bytes.NewBufferString(cl.buffer))
	tok := s.Scan()
	for tok != scanner.EOF {
		switch tok {
		case '{', '(':
			depth++
		case '}', ')':
			depth--
		}
		tok = s.Scan()
	}

	if depth < 0 {
		depth = 0
	}
	return depth
}
