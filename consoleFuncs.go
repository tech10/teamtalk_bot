package main

import (
	"errors"
	"strconv"
	"strings"
)

func console_read_prompt(prompt string) (string, error) {
	console_write(prompt + "\r\nEnter abort to cancel.")
	line, err := console_read_line()
	if strings.ToLower(line) == "abort" {
		console_write("Aborted.")
		err = errors.New("Prompt aborted.")
	}
	return line, err
}

func console_read_prompt_no_abort(prompt string) (string, error) {
	console_write(prompt)
	return console_read_line()
}

func console_read_confirm(prompt string) (bool, bool) {
	prompthead := ""
loop:
	for {
		res, err := console_read_prompt_no_abort(prompthead + "\r\n" + prompt + "Enter yes, no, or abort to cancel.")
		if err != nil {
			return false, true
		}
		switch strings.ToLower(res) {
		case "":
			prompthead = "An empty value isn't supported."
			continue loop
		case "y", "yes":
			return true, false
		case "n", "no":
			return false, false
		case "abort":
			console_write("Aborted.")
			return false, true
		default:
			prompthead = "Invalid entry."
			continue loop
		}
	}
}

func console_read_menu(prompt string, menu []string) (int, bool) {
	if len(menu) == 0 {
		return -1, true
	}
	menuselect := []string{}
	for i, string := range menu {
		if string == "" {
			continue
		}
		index := strconv.Itoa(i + 1)
		menuselect = append(menuselect, "["+index+"]: "+string)
	}
	menumsg := strings.Join(menuselect, "\r\n") + "\r\n"
	rangemin := 1
	rangemax := len(menu)
	abortmsg := "Enter abort to cancel."
	prompthead := ""
	for {
		result, err := console_read_prompt_no_abort(prompthead + prompt + menumsg + abortmsg)
		if err != nil {
			return -1, true
		}
		if strings.ToLower(result) == "abort" {
			console_write("Aborted.")
			return -1, true
		}
		if result == "" {
			prompthead = "An empty value isn't accepted.\r\n"
			continue
		}
		int, err := strconv.Atoi(result)
		if err != nil || int < rangemin || int > rangemax {
			prompthead = "Invalid selection.\r\n"
			continue
		}
		return int - 1, false
	}
}

func console_cmd() {
	defer c.wg.Done()

	for {
		line, err := console_read_line()
		if err != nil {
			if !console_closed() {
				console_close()
				quit <- true
			}
			return
		}
		if line == "" {
			continue
		}
		//Parse commands here.
		params := strings.Split(line, " ")
		if len(params) != 0 {
			cmd := strings.ToLower(params[0])
			param := ""
			if len(params) > 1 {
				param = strings.Join(params[1:], " ")
			}
			check := commands.Exists(cmd)
			if !check {
				console_write("The command " + cmd + " doesn't exist.\r\nFor a list of available commands, type \"help\".")
				continue
			}
			commands.Exec(cmd, param)
		}
	}
}
