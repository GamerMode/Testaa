package Menu

import (
	"client/ClientErrors"
	FileRequestsManager "client/FileRequests"
	HandleInput "client/HandleInput"
	"fmt"
	"net"
)

const (
	conn_addr = "clouddriveserver.duckdns.org:12345"
	prompt    = ">> "
)

type CLI struct {
	socket net.Conn
	prompt string
	input  *HandleInput.UserInput
}

func NewCLI() (*CLI, error) {
	// Connect to the server
	sock, err := net.Dial("tcp", conn_addr)
	if err != nil {
		return nil, &ClientErrors.ServerConnectionError{Err: err}
	}
	return &CLI{socket: sock, prompt: prompt, input: HandleInput.NewUserInput()}, nil
}

func (cli *CLI) closeConnection() error {
	// Close socket connection between the server
	err := cli.socket.Close()
	if err != nil {
		return err
	}
	return nil
}

// Prints the program startup intro
func (cli *CLI) PrintStartup() {
	fmt.Println("CloudDrive v1.0 Command Line Interface!")
	fmt.Println("Type \"help\" for available commands.")
}

// Print the prompt that gets output every command line
func (cli *CLI) printPrompt() {
	if FileRequestsManager.IsCurrentPathInitialized() { // If client has authenticated already
		FileRequestsManager.PrintCurrentPath() // Print the current working directory path
	}
	fmt.Print(cli.prompt)
}

func (cli *CLI) readInput() {
	cli.printPrompt()
	fmt.Println(cli.input.HandleInput(cli.socket))
}

func (cli *CLI) Loop() {
	defer cli.closeConnection()
	for {
		cli.readInput()
		if cli.input.Scanner.Bytes() == nil { // If unexpected input given
			break
		}
	}
}
