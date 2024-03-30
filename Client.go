package main

import (
	Menu "client/Menu"
	"log"
)

func main() {
	cli, err := Menu.NewCLI()
	if err != nil { // If server connection fails
		log.Fatal(err)
	}

	cli.PrintStartup()
	cli.Loop()
}
