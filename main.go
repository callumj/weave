package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args

	checkArgs(args)

	if strings.HasSuffix(args[1], ".enc") || len(args) == 3 {
		performExtraction(args)
		return
	}

	abs, err := filepath.Abs(args[1])
	if err != nil {
		log.Printf("Unable to expand %v\r\n", args[1])
		panicQuit()
	}

	performCompilation(abs)
}

func checkArgs(args []string) {
	if len(args) == 1 {
		log.Printf("Usage: %v CONFIG_FILE\r\n", args[0])
		log.Printf("Usage: %v ENCRYPTED_FILE KEY_FILE [OUT_FILE]\r\n", args[0])
		panicQuit()
	}
}

func performCompilation(configPath string) {
	fullPath := filepath.Dir(configPath)

	// ensure working dir exists
	workingDir := fmt.Sprintf("%v/working", fullPath)
	if !pathExists(workingDir) {
		log.Println("Working directory does not existing, creating")
		err := os.Mkdir(workingDir, 0775)
		if err != nil {
			log.Printf("Unable to create %v\r\n", workingDir)
			panicQuit()
		}
	}

	instr := parseInstruction(configPath)
	if instr == nil {
		panicQuit()
	}
	explainInstruction(*instr)

	baseContents := getContents(instr.Src, instr.IgnoreReg)
	if baseContents == nil {
		panicQuit()
	}
	suffix := fmt.Sprintf("%v/%v_%v.tar", workingDir, baseContents.Size, baseContents.Newest.Unix())
	baseArchive := createBaseArchive(instr.Src, baseContents.Contents, suffix)

	if baseArchive == nil {
		log.Println("Failed to create base archive.")
		panicQuit()
	}

	for _, conf := range instr.Configurations {
		thisPath := fmt.Sprintf("%v/configurations/%v", fullPath, conf.Name)
		if !pathExists(thisPath) {
			log.Printf("%v does not exist. Skipping..\r\n", thisPath)
		} else {
			log.Printf("Configuring: %v\r\n", thisPath)
			thisContents := getContents(thisPath, instr.IgnoreReg)
			tarPath := fmt.Sprintf("%v/%v_%v_%v.tar", workingDir, conf.Name, thisContents.Size, thisContents.Newest.Unix())
			if !mergeIntoBaseArchive(*baseArchive, thisPath, thisContents.Contents, tarPath) {
				log.Println("Failed to merge with base archive. Quitting.")
				panicQuit()
			}
			gzipPath := fmt.Sprintf("%v.gz", tarPath)
			compressArchive(tarPath, gzipPath)
			os.Remove(tarPath)

			if instr.Encrypt {
				cryptPath := fmt.Sprintf("%v.enc", gzipPath)
				keyFile := fmt.Sprintf("%v/keys/%v", fullPath, conf.Name)
				if !encryptFile(gzipPath, cryptPath, keyFile) {
					log.Printf("Failed to encrypt %v. Quiting..\r\n", gzipPath)
					panicQuit()
				}
				os.Remove(gzipPath)
			}
		}
	}
}

func performExtraction(args []string) {
	if len(args) < 3 {
		fmt.Printf("Usage: %v ENCRYPTED_FILE KEY_FILE [OUT_FILE]\r\n", args[0])
		panicQuit()
	}

	target := args[1]
	keyfile := args[2]

	var success bool
	if len(args) >= 4 {
		success = decryptFile(target, args[3], keyfile)
	} else {
		out := strings.Replace(target, ".enc", "", 1)
		if out == target {
			fmt.Println("Cannot determine the out file, please specify")
			panicQuit()
		}
		success = decryptFile(target, out, keyfile)
	}

	if !success {
		panicQuit()
	}
}

func panicQuit() {
	os.Exit(1)
}
