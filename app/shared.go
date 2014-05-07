package app

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Run(args []string) {
	checkArgs(args)

	if strings.HasSuffix(args[1], ".enc") || len(args) >= 3 {
		performExtraction(args)
		return
	}

	abs, err := filepath.Abs(args[1])
	if err != nil {
		panicQuitf("Unable to expand %v\r\n", args[1])
	}

	performCompilation(abs)
}

func checkArgs(args []string) {
	if len(args) == 1 {
		log.Printf("Usage: %v CONFIG_FILE\r\n", args[0])
		panicQuitf("Usage: %v ENCRYPTED_FILE KEY_FILE [OUT_FILE]\r\n", args[0])
	}
}

func panicQuit() {
	os.Exit(1)
}

func panicQuitf(format string, v ...interface{}) {
	log.Printf(format, v...)
	os.Exit(1)
}
