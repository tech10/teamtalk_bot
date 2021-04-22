package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	//Recover from a panic, which will crash the program,
	//but will recover the terminal to a sane state.
	defer func() {
		if pd := recover(); pd != nil {
			console_close()
			fmt.Fprintln(os.Stderr, pd)
			os.Exit(3)
		}
	}()
	flag.Parse()
	conf_init(cname)
	signals_init()
	conn_count := 0
	c.wg.Add(1)
	go console_cmd()
	for _, server := range c.Servers_read() {
		if autostart := server.AutoConnectOnStart_read(); autostart {
			conn_count++
			go server.Startup(autostart)
		}
	}
	if conn_count == 0 {
		console_write("Currently not connected to any servers. Use the connect command to connect to any available servers.")
	}
	c.wg.Wait()
	c.Write()
	console_write("Shutdown complete.")
}
