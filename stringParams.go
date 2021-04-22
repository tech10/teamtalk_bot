package main

import "strings"

func stringSeperateParam(str, delim, wordchar string) []string {
	params := []string{}
	if str == "" {
		return params
	}
	if delim == "" || wordchar == "" {
		return []string{str}
	}
	wordchar = string(wordchar[0])
	delim = string(delim[0])
	paramstr := ""
	word := 0
	for _, field := range strings.Split(str, delim) {
		if field == "" {
			continue
		}
		if word == 0 {
			if strings.HasPrefix(field, wordchar) && strings.HasSuffix(field, wordchar) {
				params = append(params, field)
				continue
			} else if strings.HasPrefix(field, wordchar) {
				paramstr = field + delim
				word++
				continue
			} else if strings.HasSuffix(field, wordchar) {
				params = append(params, field)
				continue
			}
			params = append(params, field)
			continue
		}
		if strings.HasPrefix(field, wordchar) {
			word++
			paramstr += field + delim
			continue
		} else if strings.HasSuffix(field, wordchar) {
			word--
			if word == 0 {
				params = append(params, paramstr+field)
				paramstr = ""
				continue
			} else {
				paramstr += field + delim
				continue
			}
		}
		paramstr += field + delim
	}
	if strings.HasSuffix(paramstr, delim) {
		if len(paramstr) > 1 {
			paramstr = paramstr[:len(paramstr)-1]
		} else {
			paramstr = ""
		}
	}
	if paramstr != "" {
		params = append(params, paramstr)
	}
	for i := range params {
		params[i] = trimParam(params[i], wordchar)
	}
	return params
}

func trimParam(str, pstr string) string {
	if strings.HasPrefix(str, pstr) {
		str = str[1:]
	}
	if strings.HasSuffix(str, pstr) {
		str = str[:len(str)-1]
	}
	return str
}

func restoreParams(params []string, delim, wordchar string) []string {
	if len(params) == 0 || delim == "" || wordchar == "" {
		return []string{}
	}
	wordchar = string(wordchar[0])
	delim = string(delim[0])
	for i, param := range params {
		if strings.Contains(param, delim) {
			params[i] = wordchar + param + wordchar
		}
	}
	return params
}
