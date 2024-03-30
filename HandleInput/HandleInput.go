package Handleinput

import (
	"bufio"
	"client/Authentication"
	FileRequestsManager "client/FileRequests"
	"net"
	"os"
	"strings"
)

const (
	prefix_index      = 0
	command_arguments = 1
)

type UserInput struct {
	Scanner *bufio.Scanner
}

func NewUserInput() *UserInput {
	return &UserInput{Scanner: bufio.NewScanner(os.Stdin)}
}

// Scan user's input and convert it to text
func (inputBuffer UserInput) readInput() string {
	inputBuffer.Scanner.Scan()
	command := inputBuffer.Scanner.Text()

	return command
}

func helpScreen() string {
	return `
SIGNUP		Create an account in CloudDrive service.
SIGNIN		Sign in to an existing CloudDrive account.
CD		Displays/Changes the current working directory.
NEWFILE		Creates a new file.
NEWDIR		Creates a new directory.
RM		Removes a content.
RENAME		Renames a folder or a directory.
MOVE		Moves a file/folder to a different location.
LS		List all the current files in the current or given path.
GARBAGE		A quick shortcut to Garbage directory.
UPLOADFILE	Uploads a file to the current directory/given directory.
DOWNLOADFILE	Downloads a file in the current program directory/given directory.
UPLOADDIR	Uploads a directory to the current directory/given directory.
DOWNLOADDIR	Downloads a directory to the current program directory/given directory.
		`
}

//Gets user input and handles its command request.

func (inputBuffer UserInput) HandleInput(socket net.Conn) string {
	var err error
	command := strings.Fields(inputBuffer.readInput())
	if len(command) > 0 { // If command is not empty
		command_prefix := strings.ToLower(command[prefix_index])

		switch command_prefix {

		case "help":
			return helpScreen()

		case "signup":
			err = Authentication.HandleSignup(command[command_arguments:], &socket)
			if err != nil {
				return err.Error()
			}

			FileRequestsManager.InitializeCurrentPath()
			return "Successfully signed up!\n"

		case "signin":
			err = Authentication.HandleSignIn(command[command_arguments:], &socket)
			if err != nil {
				return err.Error()
			}

			FileRequestsManager.InitializeCurrentPath()
			return "Successfully signed in!\n"

		case "cd":
			err = FileRequestsManager.HandleChangeDirectory(command[command_arguments:], &socket)
			if err != nil {
				return err.Error()
			}
			return ""

		case "garbage":
			err = FileRequestsManager.HandleGarbage(&socket)
			if err != nil {
				return err.Error()
			}
			return ""

		case FileRequestsManager.CreateFileCommand, FileRequestsManager.CreateFolderCommand:
			err = FileRequestsManager.HandleCreate(command, &socket)
			if err != nil {
				return err.Error()
			}
			return "The content has been created successfully!\n"

		case "rm":
			err = FileRequestsManager.HandleRemoveContent(command[command_arguments:], &socket)
			if err != nil {
				return err.Error()
			}
			return "The content has been deleted successfully!\n"

		case "rename":
			err = FileRequestsManager.HandleRename(command[command_arguments:], &socket)
			if err != nil {
				return err.Error()
			}
			return "The content has been renamed!\n"

		case "move":
			err = FileRequestsManager.HandleMove(command[command_arguments:], &socket)
			if err != nil {
				return err.Error()
			}
			return "The content has sucessfully moved!\n"

		case "ls":
			dir, err := FileRequestsManager.HandleShow(command[command_arguments:], &socket)
			if err != nil {
				return err.Error()
			}
			return dir

		case "uploadfile":
			err = FileRequestsManager.HandleUploadFile(command[command_arguments:], &socket)
			if err != nil {
				return err.Error()
			}
			return ""

		case "downloadfile":
			err = FileRequestsManager.HandleDownloadFile(command[command_arguments:], &socket)
			if err != nil {
				return err.Error()
			}
			return ""

		case "uploaddir":
			err = FileRequestsManager.HandleUploadDirectory(command[command_arguments:], &socket)
			if err != nil {
				return err.Error()
			}
			return ""

		case "downloaddir":
			err = FileRequestsManager.HandleDownloadDir(command[command_arguments:], &socket)
			if err != nil {
				return err.Error()
			}
			return ""

		default:
			return "Invalid command.\nPlease try a different command or use \"help\"\n"

		}
	}
	return ""

}
