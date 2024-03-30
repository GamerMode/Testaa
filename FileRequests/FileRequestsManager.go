package FileRequestsManager

import (
	"client/ClientErrors"
	"client/Helper"
	"client/Requests"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (
	// Command Indexes:
	/////////////////////////
	commandArgumentIndex     = 0
	pathArgumentIndex        = 0
	minimumArguments         = 1
	operationArguments       = 2
	oldFileName              = 0
	newFileName              = 1
	contentNameIndex         = 1
	remove_argument          = 1
	rename_arguments         = 2
	move_arguments           = 2
	showFolderArguments      = 1
	minimumdownloadArguments = 1
	localPathIndex           = 2
	cloudPathIndex           = 2
	/////////////////////////

	path_index = 1

	// Commands:
	CreateFileCommand   = "newfile"
	CreateFolderCommand = "newdir"
)

func convertResponeToPath(data string) string {
	parts := strings.Split(data, "CurrentDirectory:")
	return parts[path_index]
}

func HandleChangeDirectory(command_arguments []string, socket *net.Conn) error {
	if len(command_arguments) < minimumArguments { // If argument was not given
		return &ClientErrors.InvalidArgumentCountError{Arguments: uint8(len(command_arguments)), Expected: uint8(operationArguments)}
	}

	data, err := Helper.ConvertStringToBytes(strings.Join(command_arguments, " "))
	if err != nil {
		return err
	}
	responeData, err := Requests.SendRequest(Requests.ChangeDirectoryRequest, data, socket)
	if err != nil {
		return err
	}

	path := convertResponeToPath(responeData)
	setCurrentPath(path)
	return nil
}

// Handle Garbage request
func HandleGarbage(socket *net.Conn) error {
	responeData, err := Requests.SendRequest(Requests.GarbageRequest, nil, socket) // Send request type without any data

	if err != nil {
		return err
	}

	path := convertResponeToPath(responeData)
	setCurrentPath(path)
	return nil
}

// Handle create content (file or directory) requests
func HandleCreate(command []string, socket *net.Conn) error {
	if len(command) < operationArguments {
		return &ClientErrors.InvalidArgumentCountError{Arguments: uint8(len(command)), Expected: uint8(operationArguments)}
	}
	data, err := Helper.ConvertStringToBytes(strings.Join(command[contentNameIndex:], " "))
	if err != nil {
		return err
	}
	var createType Requests.RequestType

	switch command[commandArgumentIndex] {
	case CreateFileCommand:
		createType = Requests.CreateFileRequest
	case CreateFolderCommand:
		createType = Requests.CreateFolderRequest
	default:
		return fmt.Errorf("wrong create request")
	}
	_, err = Requests.SendRequest(createType, data, socket)
	return err
}

// Handle Remove Content (File and Directory)
func HandleRemoveContent(command_arguments []string, socket *net.Conn) error {
	if len(command_arguments) < remove_argument {
		return &ClientErrors.InvalidArgumentCountError{Arguments: uint8(len(command_arguments)), Expected: uint8(remove_argument)}
	}
	data, err := Helper.ConvertStringToBytes(strings.Join(command_arguments[oldFileName:], " ")) // Convert content name to raw json bytes
	if err != nil {
		return err
	}

	_, err = Requests.SendRequest(Requests.DeleteContentRequest, data, socket)
	return err
}

// Handle Rename request
func HandleRename(command_arguments []string, socket *net.Conn) error {
	if len(command_arguments) < rename_arguments { // If argument was not given
		return &ClientErrors.InvalidArgumentCountError{Arguments: uint8(len(command_arguments)), Expected: uint8(rename_arguments)}
	}
	var oldcontentName string
	var newcontentName string
	if Helper.IsQuoted(command_arguments, Helper.TwoCloudPaths) { // Check if the command arguments are enclosed within a quotation (') marks
		oldcontentName = Helper.FindPath(command_arguments, Helper.FirstNameParameter, Helper.TwoCloudPaths)
		newcontentName = fmt.Sprintf(" '" + Helper.FindPath(command_arguments, Helper.SecondNameParameter, Helper.TwoCloudPaths) + "'")
	} else {
		oldcontentName = fmt.Sprintf("'" + command_arguments[oldFileName] + "'")
		newcontentName = fmt.Sprintf(" '" + command_arguments[newFileName] + "'")
	}
	paths := oldcontentName + newcontentName // Append to a string
	data, err := Helper.ConvertStringToBytes(paths)
	if err != nil {
		return err
	}
	_, err = Requests.SendRequest(Requests.RenameRequest, data, socket)
	return err
}

// Handle Move request
func HandleMove(command_arguments []string, socket *net.Conn) error {
	if len(command_arguments) < move_arguments {
		return &ClientErrors.InvalidArgumentCountError{Arguments: uint8(len(command_arguments)), Expected: uint8(move_arguments)}
	}
	var currentFilePath string
	var newPath string
	if Helper.IsQuoted(command_arguments, Helper.TwoCloudPaths) { // Check if the command arguments are enclosed within a quotation (') marks
		currentFilePath = Helper.FindPath(command_arguments, Helper.FirstNameParameter, Helper.TwoCloudPaths)                    // Save the first path
		newPath = fmt.Sprintf(" '" + Helper.FindPath(command_arguments, Helper.SecondNameParameter, Helper.TwoCloudPaths) + "'") // Save the second path
	} else {
		currentFilePath = fmt.Sprintf("'" + command_arguments[oldFileName] + "'")
		newPath = fmt.Sprintf(" '" + command_arguments[newFileName] + "'")
	}
	paths := currentFilePath + newPath // Appened to one string
	data, err := Helper.ConvertStringToBytes(paths)
	if err != nil {
		return err
	}
	_, err = Requests.SendRequest(Requests.MoveRequest, data, socket)
	return err
}

// Handle ls command (List contents command)
func HandleShow(command_arguments []string, socket *net.Conn) (string, error) {
	if !(len(command_arguments) >= showFolderArguments || len(command_arguments) == 0) { // check for amount of arguments
		return "", &ClientErrors.InvalidArgumentCountError{Arguments: uint8(len(command_arguments)), Expected: uint8(operationArguments)}
	}
	var data []byte
	var err error
	if len(command_arguments) >= showFolderArguments { // If specific path has been specified
		data, err = Helper.ConvertStringToBytes(strings.Join(command_arguments[pathArgumentIndex:], " "))
		if err != nil {
			return "", err
		}
	} else {
		data = nil // If path hasn't been specified
	}
	respone, err := Requests.SendRequest(Requests.ShowRequest, data, socket)
	if err != nil {
		return "", err
	}
	return respone, nil
}

// Handles upload file command
func HandleUploadFile(command_arguments []string, socket *net.Conn) error {
	if len(command_arguments) < minimumArguments { // If file name was not provided
		return &ClientErrors.InvalidArgumentCountError{Arguments: uint8(len(command_arguments)), Expected: uint8(operationArguments)}
	}
	var filename string
	var cloudpath string
	if Helper.IsQuoted(command_arguments, Helper.OneClosedPath) { // Check if the first command argument is enclosed within a quotation (') marks
		filename = Helper.FindPath(command_arguments, Helper.FirstNameParameter, Helper.OneClosedPath) // Save the first path (filename to upload)
		if Helper.IsQuoted(command_arguments, Helper.TwoCloudPaths) {                                  // Check if the second command argument is enclosed within a quotation (') marks
			cloudpath = Helper.FindPath(command_arguments, Helper.SecondNameParameter, Helper.TwoCloudPaths) // Save the second path (path in cloud storage to save)
		} else { // If first path is quoted but the second doesn't
			cloudpath = Helper.ReturnNonQuotedSecondPath(command_arguments)
		}
	} else { // If command arguments are not enclosed within a quotation (') marks
		// relay on argument indexes
		filename = command_arguments[oldFileName]
		if len(command_arguments) == cloudPathIndex { // If client specificed a path to save in cloud
			cloudpath = command_arguments[newFileName]
		}
	}

	fileInfo, err := checkContent(filename) // Check if file exists, if it does returns file info api
	if err != nil {
		return err
	}
	fileSize := uint32(fileInfo.Size())
	file := newContent(filepath.Base(strings.Replace(filename, "'", "", Helper.RemoveAll)), cloudpath, fileSize) // Creates a new file struct for server communication
	file_data, err := json.Marshal(file)
	if err != nil {
		return &ClientErrors.JsonEncodeError{}
	}

	respone, err := Requests.SendRequest(Requests.UploadFileRequest, file_data, socket) // Sends upload file request
	if err != nil {                                                                     // If upload file request was rejected
		return err
	}
	chunksSize, err := Helper.ConvertResponeToChunks(respone) // Convert respone to chunks size
	if err != nil {                                           // If chunks size was returned from the server in a wrong type
		return &ClientErrors.ServerBadChunks{} // Blame the server
	}

	// Creates a privte socket connection between the server to upload the file to the server
	uploadSocket, err := Helper.CreatePrivateSocket()
	if err != nil {
		return err
	}

	go uploadFile(int64(fileSize), chunksSize, filename, true, *uploadSocket) // Upload the file with print reports

	return nil
}

// Handles download file command
func HandleDownloadFile(command_arguments []string, socket *net.Conn) error {
	if len(command_arguments) < minimumdownloadArguments {
		return &ClientErrors.InvalidArgumentCountError{Arguments: uint8(len(command_arguments)), Expected: uint8(operationArguments)}
	}
	var filename string
	var clientpath string
	if Helper.IsQuoted(command_arguments, Helper.OneClosedPath) || Helper.IsQuoted(command_arguments, Helper.TwoCloudPaths) { // Check if the first or the second command argument is enclosed within a quotation (') marks
		filename = Helper.FindPath(command_arguments, Helper.FirstNameParameter, Helper.OneClosedPath) // Save the first path (filename to upload)
		filename = filename[Helper.SkipEnclose : len(filename)-Helper.SkipEnclose]                     // Remove enclouse chars
		if Helper.IsQuoted(command_arguments, Helper.TwoCloudPaths) {                                  // Check if the second command argument is enclosed within a quotation (') marks
			clientpath = Helper.FindPath(command_arguments, Helper.SecondNameParameter, Helper.TwoCloudPaths) // Save the second path (path in cloud storage to save)
		} else { // If first path is quoted but the second doesn't
			clientpath = Helper.ReturnNonQuotedSecondPath(command_arguments)
		}
	} else { // If command arguments are not enclosed within a quotation (') marks
		// relay on argument indexes
		filename = command_arguments[oldFileName]
		if len(command_arguments) >= localPathIndex { // If local path has been specified
			clientpath = command_arguments[newFileName]
		}
	}

	// Checks if path exists
	isExists, err := Helper.IsPathExists(clientpath)
	if err != nil { // If check gone wrong
		return err
	}
	if !isExists { // If path not exists
		return &ClientErrors.PathNotExistError{Path: clientpath}
	}

	data, err := Helper.ConvertStringToBytes(filename) // Convert filename to json bytes
	if err != nil {
		return err
	}

	respone, err := Requests.SendRequest(Requests.DownloadFileRequest, data, socket) // Sends download file request
	if err != nil {
		return err
	}

	chunksSize, err := Helper.ConvertResponeToChunks(respone) // Convert respone to chunks size
	if err != nil {                                           // If chunks size was returned from the server in a wrong type
		return &ClientErrors.ServerBadChunks{} // Blame the server
	}

	fullPath := filepath.Join(clientpath, filepath.Base(filename)) // Creates the full path of the file to download

	if chunksSize == 0 { // If chunksSize that has retruend is 0, meaning the file is empty only create it.
		file, err := os.Create(fullPath) // Creates the file in the given/default path
		if err != nil {
			return &ClientErrors.CreateFileError{Filename: fullPath, Err: err}
		}
		file.Close()

		fmt.Printf("File %s has been downloaded successfully\n", fullPath)
		return nil
	}

	// Creates a privte socket connection between the server to download the file from the server
	downloadSocket, err := Helper.CreatePrivateSocket()
	if err != nil {
		return err
	}

	go downloadFile(fullPath, chunksSize, false, downloadSocket) // Start downloading file process in a seprated goroutine with success print

	return nil
}

// Handles upload directory command
func HandleUploadDirectory(command_arguments []string, socket *net.Conn) error {
	if len(command_arguments) < minimumArguments { // If dir name was not provided
		return &ClientErrors.InvalidArgumentCountError{Arguments: uint8(len(command_arguments)), Expected: uint8(operationArguments)}
	}
	var dirPath string
	var cloudpath string
	if Helper.IsQuoted(command_arguments, Helper.OneClosedPath) { // Check if the first command argument is enclosed within a quotation (') marks
		dirPath = Helper.FindPath(command_arguments, Helper.FirstNameParameter, Helper.OneClosedPath) // Save the first path (filename to upload)
		if Helper.IsQuoted(command_arguments, Helper.TwoCloudPaths) {                                 // Check if the second command argument is enclosed within a quotation (') marks
			cloudpath = Helper.FindPath(command_arguments, Helper.SecondNameParameter, Helper.TwoCloudPaths) // Save the second path (path in cloud storage to save)
		} else { // If first path is quoted but the second doesn't
			cloudpath = Helper.ReturnNonQuotedSecondPath(command_arguments)
		}
	} else { // If command arguments are not enclosed within a quotation (') marks
		// relay on argument indexes
		dirPath = command_arguments[oldFileName]
		if len(command_arguments) == cloudPathIndex { // If client specificed a path to save in cloud
			cloudpath = command_arguments[newFileName]
		}
	}
	_, err := checkContent(dirPath) // Checks if directory exists in local machine
	if err != nil {
		return err
	}
	pathSize, err := getDirSize(strings.Replace(dirPath, "'", "", Helper.RemoveAll))
	if err != nil {
		return err
	}
	dir := newContent(filepath.Base(strings.Replace(dirPath, "'", "", Helper.RemoveAll)), cloudpath, pathSize) // Creates a new dir struct for server communication
	dir_data, err := json.Marshal(dir)
	if err != nil {
		return &ClientErrors.JsonEncodeError{}
	}

	_, err = Requests.SendRequest(Requests.UploadDirectoryRequest, dir_data, socket) // Sends upload folder request
	if err != nil {                                                                  // If upload folder request was rejected
		return err
	}

	// Creates a privte socket connection between the server to upload the file to the server
	uploadSocket, err := Helper.CreatePrivateSocket()
	if err != nil {
		return err
	}
	go uploadDirectory(dirPath, *uploadSocket) // Start uploading directory process in a seprated goroutine

	return nil
}

// Handles download directory command
func HandleDownloadDir(command_arguments []string, socket *net.Conn) error {
	if len(command_arguments) < minimumdownloadArguments {
		return &ClientErrors.InvalidArgumentCountError{Arguments: uint8(len(command_arguments)), Expected: uint8(operationArguments)}
	}

	var dirname string
	var clientpath string

	if Helper.IsQuoted(command_arguments, Helper.OneClosedPath) || Helper.IsQuoted(command_arguments, Helper.TwoCloudPaths) { // Check if the first or the second command argument is enclosed within a quotation (') marks
		dirname = Helper.FindPath(command_arguments, Helper.FirstNameParameter, Helper.OneClosedPath) // Save the first path (filename to upload)
		dirname = dirname[Helper.SkipEnclose : len(dirname)-Helper.SkipEnclose]                       // Remove enclouse chars
		if Helper.IsQuoted(command_arguments, Helper.TwoCloudPaths) {                                 // Check if the second command argument is enclosed within a quotation (') marks
			clientpath = Helper.FindPath(command_arguments, Helper.SecondNameParameter, Helper.TwoCloudPaths) // Save the second path (path in cloud storage to save)
		} else { // If first path is quoted but the second doesn't
			clientpath = Helper.ReturnNonQuotedSecondPath(command_arguments)
		}
	} else { // If command arguments are not enclosed within a quotation (') marks
		// relay on argument indexes
		dirname = command_arguments[oldFileName]
		if len(command_arguments) >= localPathIndex { // If local path has been specified
			clientpath = command_arguments[newFileName]
		}
	}

	// Checks if path exists
	isExists, err := Helper.IsPathExists(clientpath)
	if err != nil { // If check gone wrong
		return err
	}
	if !isExists { // If path not exists
		return &ClientErrors.PathNotExistError{Path: clientpath}
	}

	// Checks if the directory to download is already exists in the client's PC
	isExists, err = Helper.IsPathExists(filepath.Join(clientpath, dirname))
	if err != nil {
		return err
	}
	if isExists {
		return &ClientErrors.PathExistError{Path: filepath.Join(clientpath, dirname)}
	}

	data, err := Helper.ConvertStringToBytes(dirname) // Convert filename to json bytes
	if err != nil {
		return err
	}

	_, err = Requests.SendRequest(Requests.DownloadDirRequest, data, socket) // Sends download directory request
	if err != nil {                                                          // If download directory has been rejected
		return err
	}

	// Creates a privte socket connection between the server to download the directory from the server (connects to the server)
	downloadSocket, err := Helper.CreatePrivateSocket()
	if err != nil {
		return err
	}

	go downloadDirectory(filepath.Join(clientpath, filepath.Base(dirname)), *downloadSocket) // Start downloading directory process in a seprated goroutine

	return nil
}
