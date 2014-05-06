package tools

import (
	"log"
	"os"
)

func CleanUpIfNeeded(path string) {
	if PathExists(path) {
		err := os.RemoveAll(path)
		if err != nil {
			log.Println("Failed to clean up %v\r\n", path)
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
