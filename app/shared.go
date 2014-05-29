package app

import (
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type option_set struct {
	DisableS3 bool   `short:"n" long:"no-s3" description:"Disable publishing to S3"`
	OnlyRun   string `short:"o" long:"only" description:"Only run specific confiugration"`
}

var opts option_set

func Run(args []string) {
	args, mode := checkArgs(args)

	args, err := flags.ParseArgs(&opts, args)

	if strings.HasSuffix(args[0], ".enc") || len(args) >= 2 && mode != "compile" {
		performExtraction(args)
		return
	}

	abs, err := filepath.Abs(args[0])
	if err != nil {
		panicQuitf("Unable to expand %v\r\n", args[1])
	}

	performCompilation(abs, opts)
}

func checkArgs(args []string) ([]string, string) {
	if len(args) == 1 {
		printArgMessage(args[0])
		return []string{}, ""
	}

	appname := args[0]
	args = args[1:len(args)]

	var mode string
	if args[0] == "compile" || args[0] == "extract" {
		if len(args) == 1 {
			printArgMessage(appname)
			return []string{}, ""
		} else {
			mode = args[0]
			args = args[1:len(args)]
		}
	}

	return args, mode
}

func printArgMessage(appname string) {
	log.Printf("Usage: %v CONFIG_FILE\r\n", appname)
	panicQuitf("Usage: %v ENCRYPTED_FILE KEY_FILE [OUT_FILE]\r\n", appname)
}

func panicQuit() {
	os.Exit(1)
}

func panicQuitf(format string, v ...interface{}) {
	log.Printf(format, v...)
	os.Exit(1)
}
