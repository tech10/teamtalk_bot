package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type Commands struct {
	cmd      map[string]func(string)
	cmdorder []string
	cmdhelp  map[string][]string
}

func (commands *Commands) Add(cmd string, f func(string)) {
	if commands.Exists(cmd) {
		return
	}
	commands.cmd[cmd] = f
	commands.cmdorder = append(commands.cmdorder, cmd)
	sort.Strings(commands.cmdorder)
}

func (commands *Commands) AddHelp(cmd string, help ...string) {
	if len(help) == 0 {
		return
	}
	if commands.ExistsHelp(cmd) {
		return
	}
	commands.cmdhelp[cmd] = help
}

func (commands *Commands) Exists(cmd string) bool {
	_, exists := commands.cmd[cmd]
	return exists
}

func (commands *Commands) ExistsHelp(cmd string) bool {
	_, exists := commands.cmdhelp[cmd]
	return exists
}

func (commands *Commands) HelpText(cmd string) string {
	if !commands.ExistsHelp(cmd) {
		return ""
	}
	desc := commands.cmdhelp[cmd][0]
	examples := ""
	if len(commands.cmdhelp[cmd]) > 1 {
		examples = strings.Join(commands.cmdhelp[cmd][1:], "\r\n")
	}
	helpmsg := []string{"Command: " + cmd, "Description: " + desc}
	if examples != "" {
		helpmsg = append(helpmsg, "Examples:", examples)
	}
	if len(helpmsg) == 1 {
		return helpmsg[0]
	}
	return strings.Join(helpmsg, "\r\n")
}

func (commands *Commands) Exec(cmd, param string) bool {
	if !commands.Exists(cmd) {
		return false
	}
	defer func() {
		if pd := recover(); pd != nil {
			fmt.Fprintln(os.Stderr, "PANIC", pd)
		}
	}()
	commands.cmd[cmd](param)
	return true
}

func NewCommands() Commands {
	return Commands{
		cmd:      make(map[string]func(string)),
		cmdorder: make([]string, 0),
		cmdhelp:  make(map[string][]string),
	}
}
