package main

const TT_BEEP string = "\a"

func str_yes_no(yesno bool) string {
	if yesno {
		return "yes"
	}
	return "no"
}

func stringFormatWithBS(str string) string {
	if str == "" {
		return ""
	}
	ts := ""
	for _, chr := range str {
		if chr == 8 {
			if ts != "" && len(ts) > 1 {
				ts = ts[:len(ts)-1]
				continue
			} else if len(ts) == 1 {
				ts = ""
				continue
			}
			continue
		}
		ts += string(chr)
	}
	return ts
}
