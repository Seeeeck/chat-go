package main

import (
	"flag"
	"fmt"
)

var serverIP string
var serverPort int

func init() {
	flag.StringVar(&serverIP, "ip", "127.0.0.1", "Set the IP to connect.")
	flag.IntVar(&serverPort, "p", 8888, "Set the port.")
}

func main() {
	flag.Parse()
	client := NewClient(serverIP, serverPort)
	if client == nil {
		fmt.Println("Connect fail")
		return
	}
	client.Run()
}
