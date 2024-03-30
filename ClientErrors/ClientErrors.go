package ClientErrors

import "fmt"

type SendDataError struct{ Err error }
type ReciveDataError struct{ Err error }
type ServerConnectionError struct{ Err error }
type JsonEncodeError struct{ Err error }
type JsonDecodeError struct{ Err error }
type FileNotExistError struct{ Filename string }
type PathNotExistError struct{ Path string }
type PathExistError struct{ Path string }
type ReadFileInfoError struct{ Filename string }
type ServerBadChunks struct{}
type BadFileContent struct{ Filename string }
type TimeOutRespone struct{}
type ConvertToRelative struct{}

type CreateFolderError struct {
	Foldername string
	Err        error
}

type CreateFileError struct {
	Filename string
	Err      error
}

type InvalidArgumentCountError struct {
	Arguments uint8
	Expected  uint8
}

func (error *ReciveDataError) Error() string {
	return fmt.Sprintf("error when reciving a response from the server.\n%s", error.Err)
}

func (error *SendDataError) Error() string {
	return fmt.Sprintf("Error when attempting to send the data to the server.\n%s", error.Err)
}

func (error *ServerConnectionError) Error() string {
	return fmt.Sprintf("There has been an error connecting to the server.\nPlease check your connection and try again.\nIf it doesn't work contact the developers and send them this error message:\n\n%s", error.Err)
}

func (error *JsonDecodeError) Error() string {
	return fmt.Sprintf("There has been an Error when attempting to decode the response from the server.\nPlease send this info to the developers:\n%s", error.Err)
}

func (error *JsonEncodeError) Error() string {
	return fmt.Sprintf("There has been an error when attempting to encode the data to be sent to the server.\nPlease send this info to the developers:\n%s", error.Err)
}

func (error *InvalidArgumentCountError) Error() string {
	return fmt.Sprintf("Incorrect number of arguments. got %d, expected %d arguments\nPlease try again", error.Arguments, error.Expected)
}

func (error *FileNotExistError) Error() string {
	return fmt.Sprintf("File '%s' does not exist on your local machine.", error.Filename)
}

func (error *PathNotExistError) Error() string {
	return fmt.Sprintf("Path '%s' does not exist on your local machine.", error.Path)
}

func (error *PathExistError) Error() string {
	return fmt.Sprintf("Path '%s' is already exist on your local machine.", error.Path)
}

func (error *ReadFileInfoError) Error() string {
	return fmt.Sprintf("Cannot read file %s info.", error.Filename)
}

// Usage: Upload directory process
func (error *ConvertToRelative) Error() string {
	return "error converting path to relative.\nUploading process has stopped"
}

func (error *ServerBadChunks) Error() string {
	return "Server has returned wrong type of chunks. Please contact the developers"
}

func (error *BadFileContent) Error() string {
	return fmt.Sprintf("There is a problem uploading the file content of the provided file: %s", error.Filename)
}

func (error *TimeOutRespone) Error() string {
	return "It took too long time to get a respone back from the server."
}

func (error *CreateFolderError) Error() string {
	return fmt.Sprintf("Couldn't create directory '%s'. The folder and its sub-contents might won't be available.\nPlease provide the next error information to the developers:\n%s", error.Foldername, error.Err)
}

func (error *CreateFileError) Error() string {
	return fmt.Sprintf("Couldn't create file '%s'. The file won't be available.\nPlease provide the next error information to the developers:\n%s", error.Filename, error.Err)
}
