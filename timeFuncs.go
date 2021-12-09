package main

import (
	"time"

	"github.com/hako/durafmt"
)

func time_duration_str(duration time.Duration) string {
	str := durafmt.Parse(duration).String()
	if str == "" {
		str = "0 seconds"
	}
	return str
}
