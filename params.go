package main

import (
	"regexp"
	"strconv"
	"strings"
)

var pexp *regexp.Regexp

func init() {
	paramname := `([a-zA-Z0-9._-]+)`
	digit := `(-?\d+)`
	str := `"(([^\\^"]+|\\n|\\r|\\"|\\\\)*)"`
	list := `\[([-?\d+]?[,-?\d+]*)\]`
	pexp = regexp.MustCompile(`^` + paramname + `=(` + list + `|` + digit + `|` + str + `)`)
}

func teamtalk_get_cmd(str string) string {
	params := strings.Split(str, " ")
	if len(params) != 0 {
		return params[0]
	}
	return ""
}

func teamtalk_get_params(cmdline string) map[string]string {
	rparams := make(map[string]string)
	cmdlines := strings.Split(cmdline, " ")
	if len(cmdlines) > 1 {
		cmdline = strings.Join(cmdlines[1:], " ")
	} else {
		return rparams
	}
	for {
		matches := pexp.FindStringSubmatch(cmdline)
		if len(matches) > 0 {
			pname := matches[1]
			param := matches[2]
			if strings.HasPrefix(param, `"`) {
				param = param[1:]
			}
			if strings.HasSuffix(param, `"`) {
				param = param[:len(param)-1]
			}
			rparams[pname] = teamtalk_format_from_param(param)
			if len(cmdline) > len(matches[0]) {
				cmdline = cmdline[len(matches[0])+1:]
			} else {
				cmdline = ""
				break
			}
		}
	}
	return rparams
}

func teamtalk_param_find(params map[string]string, param string) bool {
	if param == "" {
		return false
	}
	_, exists := params[param]
	return exists
}

func teamtalk_param_int(params map[string]string, param string) (int, bool) {
	if !teamtalk_param_find(params, param) {
		return 0, false
	}
	int, err := strconv.Atoi(params[param])
	if err != nil {
		return 0, false
	}
	return int, true
}

func teamtalk_param_str(params map[string]string, param string) string {
	if !teamtalk_param_find(params, param) {
		return ""
	}
	return params[param]
}

func teamtalk_param_list(params map[string]string, param string) []int {
	if !teamtalk_param_find(params, param) {
		return []int{}
	}
	if !strings.HasPrefix(params[param], "[") && !strings.HasSuffix(params[param], "]") {
		return []int{}
	}
	data := strings.Split(params[param][1:len(params[param])-1], ",")
	ints := []int{}
	for _, int := range data {
		if int == "" {
			continue
		}
		num, err := strconv.Atoi(int)
		if err != nil {
			continue
		}
		ints = append(ints, num)
	}
	return ints
}

func teamtalk_format_list(ints []int) string {
	if len(ints) == 0 {
		return "[]"
	}
	list := []string{}
	for _, int := range ints {
		list = append(list, strconv.Itoa(int))
	}
	return "[" + strings.Join(list, ",") + "]"
}

func teamtalk_format_cmd(cmd string, params ...string) string {
	cmdjoin := []string{cmd}
	paramchar := "="
	strchar := "\""
	liststartchar := "["
	listendchar := "]"
	if len(params) > 1 {
		for i := 0; i < len(params)-1; i += 2 {
			param1 := params[i]
			param2 := params[i+1]
			_, err := strconv.Atoi(param2)
			if err == nil {
				cmdjoin = append(cmdjoin, param1+paramchar+param2)
				continue
			}
			if strings.HasPrefix(param2, liststartchar) && strings.HasSuffix(param2, listendchar) {
				cmdjoin = append(cmdjoin, param1+paramchar+param2)
				continue
			}
			param2 = teamtalk_format_to_param(param2)
			cmdjoin = append(cmdjoin, param1+paramchar+strchar+param2+strchar)
		}
	}
	return strings.Join(cmdjoin, " ")
}

func teamtalk_format_to_param(str string) string {
	str = strings.Replace(str, `\`, `\\`, -1)
	str = strings.Replace(str, `"`, `\"`, -1)
	str = strings.Replace(str, "\r", "\\r", -1)
	str = strings.Replace(str, "\n", "\\n", -1)
	return str
}

func teamtalk_format_from_param(str string) string {
	str = strings.Replace(str, "\\r", "\r", -1)
	str = strings.Replace(str, "\\n", "\n", -1)
	str = strings.Replace(str, `\"`, `"`, -1)
	str = strings.Replace(str, `\\`, `\`, -1)
	return str
}
