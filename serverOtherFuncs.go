package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

//Function to ping all servers here.

func servers_ping(ping_times int) {
	defer func() {
		pd := recover()
		if pd != nil {
			fmt.Fprintln(os.Stdout, "PANIC:", pd)
		}
	}()
	if len(c.Servers_read()) == 0 {
		return
	}
	console_write("Pinging all connected servers. This may take a while.\r\nOther commands can be used during server pinging.")
	t_start := time.Now()
	msecs_min := 0
	msecs_max := 0
	msecs_avg := 0
	var server_min *tt_server
	var server_max *tt_server
	count := 0
	for _, server := range c.Servers_read() {
		if !server.connected() {
			continue
		}
		count++
		for i := 1; i <= ping_times; i++ {
			start := time.Now()
			server.cmd_ping()
			end := time.Since(start)
			msecs := int(end.Nanoseconds()) / int(time.Millisecond) / 3
			msecs_avg += msecs
			if i == 1 {
				msecs_min = msecs
				server_min = server
				msecs_max = msecs
				server_max = server
				continue
			}
			if msecs > msecs_max {
				msecs_max = msecs
				server_max = server
			}
			if msecs < msecs_min {
				msecs_min = msecs
				server_min = server
			}
		}
	}
	if count == 0 {
		console_write("No servers were available to ping.")
		return
	}
	t_end := time.Since(t_start)
	t_msecs := int(t_end.Nanoseconds()) / int(time.Millisecond)
	msecs_total := msecs_avg
	msecs_avg = int(msecs_avg / (ping_times * count))
	msg := "Ping complete. " + strconv.Itoa(count) + " server"
	if count != 1 {
		msg += "s"
	}
	msg += " have been pinged.\r\nMinimum milliseconds: " + strconv.Itoa(msecs_min)
	if server_min != nil {
		msg += " (" + server_min.DisplayName_read() + ")"
	}

	msg += "\r\nMaximum milliseconds: " + strconv.Itoa(msecs_max)
	if server_max != nil {
		msg += " (" + server_max.DisplayName_read() + ")"
	}

	msg += "\r\nAverage milliseconds: " + strconv.Itoa(msecs_avg) + "\r\nTotal milliseconds: " + strconv.Itoa(msecs_total) + "\r\nTotal milliseconds taken to ping available servers: " + strconv.Itoa(t_msecs)
	console_write(msg)
}
