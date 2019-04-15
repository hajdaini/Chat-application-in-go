package main

import (
	"app/chat_app/client"
	"app/chat_app/server"
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	IP   = "127.0.0.1"
	PORT = "3569"
)

func usage() {
	fmt.Print("Unrecognized option \n\n")
	fmt.Println("Usage : go run main.go --mode <mode>")
	fmt.Print("- mode : \"server\" or \"client \n\n")
	fmt.Println("example to run a server: go run main.go --mode server")
	fmt.Println("example to run a client: go run main.go --mode client")
}

func options() {
	var mode string

	// by default it will run client mode
	flag.StringVar(&mode, "mode", "client", "--mode client or --mode server")
	flag.Parse()

	if strings.ToLower(mode) == "server" {
		server := server.New(IP, PORT)
		server.Run()
	} else if strings.ToLower(mode) == "client" {
		client := client.New(IP, PORT)
		client.Run()
	} else {
		usage()
		os.Exit(2)
	}
}

func main() {
	options()
}
