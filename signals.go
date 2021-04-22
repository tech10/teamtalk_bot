package main

import (
	"os"
	"os/signal"
	"syscall"
)

var quit chan bool

func signals_init() {
	//Signal notifiers.
	kill := make(chan os.Signal)
	signal.Notify(kill,
		os.Interrupt,
		os.Kill,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	quit = make(chan bool)
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer func() {
			recover()
		}()
	loop:
		for {
			select {
			case <-kill:
				console_write("Kill signal received.")
				console_close()
				break loop
			case <-quit:
				console_write("Locally initiated shutdown received.")
				break loop
			}
		}
		shutdown()
		if !console_closed() {
			console_close()
		}
	}()
}

func shutdown() {
	for _, server := range c.Servers_read() {
		server.Shutdown()
	}
}
