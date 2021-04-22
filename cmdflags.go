package main

import "flag"

var cname string
var wd string

func init() {
	flag.StringVar(&cname, "c", "config.xml", "Name or full path to configuration file.")
	flag.StringVar(&wd, "d", "", "Working directory.")
}
