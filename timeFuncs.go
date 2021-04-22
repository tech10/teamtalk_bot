package main

import (
	"github.com/hako/durafmt"
	"time"
)

func time_duration_str(duration time.Duration) string {
	str := durafmt.Parse(duration).String()
	if str == "" {
		str = "0 seconds"
	}
	return str
}
