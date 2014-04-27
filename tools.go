package main

import (
	"log"
	"os"
)

func cleanUpIfNeeded(path string) {
	if pathExists(path) {
		err := os.RemoveAll(path)
		if err != nil {
			log.Println("Failed to clean up %v\r\n", path)
		}
	}
}

func pathExists(path string) bool {
	_, checkErr := os.Stat(path)
	if checkErr != nil && os.IsNotExist(checkErr) {
		return false
	} else {
		return true
	}
}
