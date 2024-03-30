package FileRequestsManager

import (
	"bufio"
	"client/ClientErrors"
	"client/Helper"
	"client/Requests"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	empty    = 0
	kilobyte = 1_000_000

	firstIndex = 0
	chunksSize = 11

	stopTransmissionRespone = "{\"Type\":501,\"Data\":\"\"}"
)

//var mu sync.Mutex // Lock the file writing to make sure only one goroutine can write over the file

// Uploads file to the cloud server with the given file size an the chunk from the cloud core technology, filename to open the file from local pc
// shoutFlag to indicate whether to print upload finished or no
func uploadFile(fileSize int64, chunksSize int, filename string, shoutFlag bool, socket net.Conn) {
	file, err := os.Open(strings.Replace(filename, "'", "", Helper.RemoveAll)) // Open file
	if err != nil {
		fmt.Println(err.Error())
	}
	defer file.Close()

	chunk := make([]byte, chunksSize) // Save buffer of chunks

	var validUpload = true
	var totalBytesRead int64
	var totalReadFlag int64 // Flag for client view to automatically update upload percentage and bar progress
	var precentage int64

	for {
		bytesRead, err := file.Read(chunk)
		if err == io.EOF { // If finish reading file succesfully
			break
		}
		if err != nil { // If error occurred while reading the file
			validUpload = false
			fmt.Println(err.Error())
			err = &ClientErrors.ReadFileInfoError{}
			fmt.Println(err.Error())
		}
		if bytesRead == empty { // If finish reading file succesfully
			break
		}

		_, err = socket.Write(chunk[:bytesRead]) // Sending chunk to server
		if err != nil {                          // If sending error occured
			validUpload = false
			err = &ClientErrors.SendDataError{}
			fmt.Println(err.Error())
		}

		// Add bytes read for the chunk to the flag percentage and to the total read bytes
		totalReadFlag += int64(bytesRead)
		totalBytesRead += int64(bytesRead)

		if totalReadFlag >= kilobyte && shoutFlag { // If shout flag has been enabled, for every 1 Kilobyte update the progess and perecntage bar in the cli
			totalReadFlag = 0
			precentage = (totalBytesRead * 100) / fileSize // Calculates total read bytes compared to the total file size in percentages
			printer := func(_ int, _ string) string {
				var bar string
				for i := 0; i < int(precentage/2); i++ {
					bar += "-"
				}
				return bar
			}
			fmt.Printf("\033[F\033[K")
			fmt.Printf("Upload Progress: %v%% - %s", precentage, printer(int(precentage), "-"))
			fmt.Println()

		}
	}
	// If upload finished and shout flag has been enabled
	if validUpload && shoutFlag {
		fmt.Printf("File %s has been uploaded successfully\n", filename)
	}
}

// Upload directory to cloud server
func uploadDirectory(dirpath string, socket net.Conn) {
	err := filepath.WalkDir(dirpath, func(contentPath string, contentInfo fs.DirEntry, err error) error { // Walk through all the contents in the given dir path
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(dirpath, contentPath) // Convert the contentPath absolute to relative from the given path to upload
		if err != nil {
			return &ClientErrors.ConvertToRelative{}
		}
		if relativePath != "." { // If path is not the base (already exists) path
			dirData, err := Helper.ConvertStringToBytes(relativePath) // Convert new dir path to bytes
			if err != nil {
				return err
			}

			if !contentInfo.IsDir() { // If content is file

				fileInfo, err := contentInfo.Info() // Get file's info
				if err != nil {
					return &ClientErrors.ReadFileInfoError{Filename: filepath.Base(relativePath)}
				}

				// Initializes file struct
				file := newContent(filepath.Base(relativePath), filepath.Dir(relativePath), uint32(fileInfo.Size()))
				// Convert file struct to json bytes
				file_data, err := json.Marshal(file)
				if err != nil {
					return &ClientErrors.JsonEncodeError{Err: err}
				}

				// Sends Upload File reques
				respone, err := Requests.SendRequest(Requests.UploadFileRequest, file_data, &socket)
				if err != nil { // If upload file request was rejected
					return err
				}

				chunksSize, err := Helper.ConvertResponeToChunks(respone) // Convert respone to chunks size
				if err != nil {                                           // If chunks size was returned from the server in a wrong type
					return &ClientErrors.ServerBadChunks{} // Blame the server
				}

				uploadFile(fileInfo.Size(), chunksSize, contentPath, false, socket) // Uploads the file with no prints

			} else { // If content is directory
				// Sends request to make a new directory
				respone, err := Requests.SendRequestInfo(Requests.BuildRequestInfo(Requests.CreateFolderRequest, dirData), true, socket)
				if err != nil {
					return err
				}

				if respone.Type == Requests.ErrorRespone { // If respone is error
					return fmt.Errorf(respone.Respone)
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	_, err = Requests.SendRequestInfo(Requests.BuildRequestInfo(Requests.StopTransmission, nil), false, socket) // Send stop upload request to server
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("Upload directory has finished")

}

// Write the file content on a seprated goroutine to not waste time and resources for the main thread that recives the file
// func writeFile(writer **bufio.Writer, chunkBytes []byte) {
// 	fmt.Println("Locking... (Currently Unlocked)")
// 	mu.Lock()         // Acquire lock before writing the file, promising that only one goroutine can write over the file
// 	defer mu.Unlock() // Ensure the lock is released after writing the file
// 	fmt.Println("It's locked")
// 	_, err := (*writer).Write(chunkBytes)
// 	if err != nil {
// 		fmt.Println("Error writing data on the provided file path.\nPlease contact the developers")
// 		return
// 	}
// 	fmt.Println("Unlocking... (Currently locked)")
// }

// TDL add file's size to the argument so it would print percentage bar
// Download a file from the cloud server, prints finished downloading file if supression flag is off, prints errors in any case.
func downloadFile(path string, chunksSize int, suppression bool, socket *net.Conn) {
	file, err := os.Create(path) // Creates the file in the given/default path
	if err != nil {
		fmt.Println("Couldn't create the file in the provided path.\nPlease provide a different path.")
		return
	}
	file.Close()

	file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644) // Open file for writing
	if err != nil {
		fmt.Println("Couldn't open the file for writing data.\nPlease make sure you have provided a path with permissions for write.")
		return
	}
	defer file.Close()

	// Create a buffered writier for efficient writes
	writer := bufio.NewWriter(file)

	for {
		chunkBytes, err := Helper.ReciveChunkData(socket, chunksSize)
		// If the client hasn't recived any new chunks for over the configured timeout, finish reading file sucessfully
		if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() || string(chunkBytes) == stopTransmissionRespone {
			break
		}
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		_, err = (*writer).Write(chunkBytes)
		if err != nil {
			fmt.Println("Error writing data on the provided file path.\nPlease contact the developers")
			return
		}
		//go writeFile(&writer, chunkBytes) // Write the file over goroutine to not interept the connection
	}
	err = writer.Flush() // Flush any remaining data in the buffer to the file
	if err != nil {
		fmt.Println("Error flushing data to the file.\nPlease contact the developers")
		return
	}
	// If suppression flag is off, prints success
	if !suppression {
		fmt.Printf("File %s has been downloaded successfully\n", path)
	}
}

func createFolder(info Requests.ResponeInfo, baseFolderPath string) error {
	fullPath := filepath.Join(baseFolderPath, info.Respone) // Append the full path of the new directory by the base Folder path and the given folder path
	err := os.Mkdir(fullPath, os.ModePerm)                  // Creates a folder
	if err != nil {                                         // If creating folder was unsuccessfull
		return &ClientErrors.CreateFolderError{Foldername: fullPath, Err: err}
	}
	return nil
}

// Implement getFileInfo for reciving folder implemention.
// Returns all the file's info that is given from the server.
// Input:
// Requests.ResponeInfo - Server's Respone struct
// baseFolderPath - Base path of client side
// Output:
// File's chunks size (If file's valid)
// File's size (If file's valid)
// Absolute filepath (if file's valid)
// error (if file's not valid)
func getFileInfo(socket *net.Conn, info Requests.ResponeInfo, baseFolderPath string) (uint32, uint32, string, error) {
	content, err := ParseDataToContent(info.Respone) // Convert string json respone to content struct
	if err != nil {
		return empty, empty, "", err
	}

	absFilePath := filepath.Join(baseFolderPath, content.Path, content.Name) // Convert to absolute file path

	dataBytes, err := Helper.ReciveData(socket) // Recieves chunks size bytes json data from server
	if err != nil {
		return empty, empty, "", err
	}
	responeInfo, err := Requests.GetResponseInfo(dataBytes) // Convert raw bytes json to ResponeInfo struct
	if err != nil {
		return empty, empty, "", err
	}
	if responeInfo.Type != Requests.ValidRespone { // If respone valid chunks hasn't recieved
		return empty, empty, "", fmt.Errorf(responeInfo.Respone) // Returns error with its error data
	}
	chunks, err := strconv.ParseUint(responeInfo.Respone[chunksSize:], 10, 32)
	if err != nil {
		return empty, empty, "", err
	}

	return uint32(chunks), content.Size, absFilePath, nil
}

func downloadDirectory(path string, socket net.Conn) {
	os.Mkdir(path, os.ModePerm) // Creates the base directory with set permissions for the directory

	var startDownload = func() error {
		// Start reciving contents in the base directory
		for {
			dataBytes, err := Helper.ReciveData(&socket) // Recieves bytes json data from server
			if err != nil {
				return err
			}
			responeInfo, err := Requests.GetResponseInfo(dataBytes) // Convert raw bytes json to ResponeInfo struct
			if err != nil {
				return err
			}
			// ResponeInfo is like RequestInfo, receiving RequestInfo types in ResponeInfo struct

			switch responeInfo.Type {
			case Requests.ResponeType(Requests.CreateFolderRequest):
				// If server pointed at a directory to create
				err = createFolder(responeInfo, path)
				if err != nil {
					fmt.Println(err.Error()) // Print create folder error, so it won't stop the reciving folder proccess
				}
			case Requests.ResponeType(Requests.DownloadFileRequest):
				// If server pointed at a file to recieve
				chunkSize, fileSize, fileAbsPath, err := getFileInfo(&socket, responeInfo, path) // Get all file's info by its ResponeInfo detail
				if err != nil {
					return err
				}
				// Avoid downloading empty file
				if fileSize > 0 {
					downloadFile(fileAbsPath, int(chunkSize), true, &socket) // Start downloading file proccess with no success prints
				} else {
					// If file is empty, only create it
					file, err := os.Create(fileAbsPath) // Creates the file in the given/default path
					if err != nil {
						return &ClientErrors.CreateFileError{Filename: fileAbsPath, Err: err}
					}
					file.Close()
				}

			case Requests.ResponeType(Requests.StopTransmission): // If server indicated that the download proccess is finished
				return nil
			}

		}
	}
	err := startDownload() // Start downloading proccess
	if err != nil {
		switch err := err.(type) {
		case *net.OpError:
			if err.Timeout() { // If error type is Timeout
				fmt.Println("Downloading might have finished. Could not verify the download with the server.\nPlease make sure all the contents have been successfully downloaded.")
			}
		default: // If it's any other error type
			fmt.Println(err.Error() + "\nDownload process has been stopped.")
		}
		return
	}
	fmt.Println("Finished Downloading " + path + " Path")
}
