package tools

import (
	"log"
	"os"
)

func CleanUpIfNeeded(path string) {
	if PathExists(path) {
		err := os.RemoveAll(path)
		if err != nil {
			log.Printf("Failed to clean up %s\r\n", path)
		}
	}
}

func PathExists(path string) bool {
	_, checkErr := os.Stat(path)
	if checkErr != nil && os.IsNotExist(checkErr) {
		return false
	} else {
		return true
	}
}
