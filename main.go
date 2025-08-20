package main

import (
	"flag"
	"fmt"
	"zabbix-source/config"
)

var (
	cPath = flag.String("c", "", "relative path to config file")
)

func main() {
	flag.Parse()
	if *cPath == "" {
		fmt.Println("config file path is required")
		return
	}
	_, err := config.Parse(*cPath)
	if err != nil {
		fmt.Println("failed to parse config file:", err)
	}
}
