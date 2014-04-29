package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"sort"
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

func GenerateNameSuffix(info ContentsInfo) string {
	sortedNames := info.Contents
	sort.Strings(sortedNames)

	var buffer bytes.Buffer
	for _, name := range sortedNames {
		buffer.WriteString(name)
	}

	hasher := sha256.New()
	hasher.Write(buffer.Bytes())
	hashed := hasher.Sum(nil)
	hashedString := hex.EncodeToString(hashed)
	return fmt.Sprintf("%s_%v", hashedString, info.Size)
}
