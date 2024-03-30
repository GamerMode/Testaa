package Helper

import "os"

// Returns whether the given path exists. Returns error if the check gone wrong
func IsPathExists(path string) (bool, error) {
	if path != "" { // If path has any value, otherwise returns true automatically
		_, err := os.Stat(path)
		if err == nil { // If path exists
			return true, nil
		}
		if os.IsNotExist(err) { // If path not exists
			return false, nil
		}
		// If the check gone wrong
		return false, err
	}
	return true, nil
}
