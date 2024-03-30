package FileRequestsManager

import "fmt"

var (
	CurrentPath string
)

func InitializeCurrentPath() {
	CurrentPath = "Root:\\"
}

func PrintCurrentPath() {
	fmt.Print(CurrentPath)
}

func IsCurrentPathInitialized() bool {
	return CurrentPath != ""
}

func setCurrentPath(path string) {
	CurrentPath = path
}
