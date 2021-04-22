package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/chzyer/readline"
	"os"
	"runtime"
	"strings"
	"sync"
)

var rl *readline.Instance
var lrl sync.Mutex

func init() {
	console_open()
}

func console_open() bool {
	var err error
	if !console_use_readline() {
		console_write("WARNING: Readline unavailable. Command history may be unavailable. Do not use arrow keys.")
		return true
	}
	rl, err = readline.NewEx(&readline.Config{
		UniqueEditLine: false,
	})
	if err != nil {
		fmt.Println("WARNING: Failed to open go readline console:", err, "Fallback to standard go console.")
		rl = nil
		return false
	}
	rl.SetPrompt("> ")
	return true
}

func console_clear() {
	defer lrl.Unlock()
	lrl.Lock()
	for i := 0; i < 500; i++ {
		if rl != nil {
			fmt.Fprintln(rl, "")
		} else {
			fmt.Println("")
		}
	}
}

func console_use_readline() bool {
	ops := runtime.GOOS
	switch ops {
	case "windows":
		return false
	}
	return true
}

func console_read_line() (string, error) {
	var err error
	var line string
	if rl != nil && console_use_readline() {
		res := rl.Line()
		if res.CanBreak() || res.Error != nil {
			line = ""
			err = errors.New("Input interrupted.")
		} else {
			line = strings.Trim(res.Line, "\r\n ")
			err = nil
		}
	} else {
		line, err = bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", err
		}
		line = stringFormatWithBS(strings.Trim(line, "\r\n "))
	}
	return line, err
}

func console_write(data string) {
	data = strings.TrimSuffix(data, "\r\n")
	if data == "" {
		return
	}
	console_writec(data)
}

func console_writec(data string) {
	defer lrl.Unlock()
	lrl.Lock()
	if rl == nil {
		fmt.Println(data)
	} else {
		fmt.Fprintln(rl, data)
	}
}

func console_close() {
	defer func() {
		recover()
	}()
	defer lrl.Unlock()
	lrl.Lock()
	if rl != nil {
		rl.Close()
		rl = nil
	}
	os.Stdin.Close()
}

func console_closed() bool {
	if console_use_readline() {
		if rl == nil {
			return true
		} else {
			return false
		}
	}
	return true
}
