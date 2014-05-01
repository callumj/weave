package main

import (
	"callumj.com/weave/core"
	"callumj.com/weave/upload"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	args := os.Args

	checkArgs(args)

	if strings.HasSuffix(args[1], ".enc") || len(args) >= 3 {
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
	if !core.PathExists(workingDir) {
		log.Println("Working directory does not existing, creating")
		err := os.Mkdir(workingDir, 0775)
		if err != nil {
			log.Printf("Unable to create %v\r\n", workingDir)
			panicQuit()
		}
	}

	instr := core.ParseInstruction(configPath)
	if instr == nil {
		panicQuit()
	}
	core.ExplainInstruction(*instr)

	baseContents := core.GetContents(instr.Src, instr.IgnoreReg)
	if baseContents == nil {
		panicQuit()
	}

	baseSuffix := core.GenerateNameSuffix(*baseContents)
	baseFileName := fmt.Sprintf("%v/%v.tar", workingDir, baseSuffix)
	baseArchive := core.CreateBaseArchive(instr.Src, baseContents.Contents, baseFileName)

	if baseArchive == nil {
		log.Println("Failed to create base archive.")
		panicQuit()
	}

	var col []upload.FileDescriptor

	for _, conf := range instr.Configurations {
		thisPath := fmt.Sprintf("%v/configurations/%v", fullPath, conf.Name)
		log.Printf("Configuring: %v\r\n", thisPath)
		var thisContents *core.ContentsInfo
		if core.PathExists(thisPath) {
			thisContents = core.GetContents(thisPath, instr.IgnoreReg)
		} else {
			thisContents = new(core.ContentsInfo)
			thisContents.Size = 0
			thisContents.Contents = []core.FileInfo{}
			thisContents.Newest = baseContents.Newest
		}

		filteredContents := core.FilterContents(*baseContents, conf.ExceptReg, conf.OnlyReg)
		recalcBaseSuffix := core.GenerateNameSuffix(*filteredContents)
		tarPath := fmt.Sprintf("%v/%v_%v.tar", workingDir, conf.Name, core.GenerateFinalNameSuffix(recalcBaseSuffix, *thisContents))

		if !core.MergeIntoBaseArchive(*baseArchive, thisPath, thisContents.Contents, tarPath, filteredContents) {
			log.Println("Failed to merge with base archive. Quitting.")
			panicQuit()
		}
		gzipPath := fmt.Sprintf("%v.gz", tarPath)
		core.CompressArchive(tarPath, gzipPath)
		os.Remove(tarPath)

		finalPath := gzipPath
		if instr.Encrypt {
			cryptPath := fmt.Sprintf("%v.enc", gzipPath)
			keyFile := fmt.Sprintf("%v/keys/%v", fullPath, conf.Name)
			if !core.EncryptFile(gzipPath, cryptPath, keyFile) {
				log.Printf("Failed to encrypt %v. Quiting..\r\n", gzipPath)
				panicQuit()
			} else {
				finalPath = cryptPath
			}
			os.Remove(gzipPath)
		}

		if instr.S3 != nil {
			stat, err := os.Stat(finalPath)
			if err != nil {
				log.Println("Unable to query %v\r\n", finalPath)
				panicQuit()
			}

			desc := new(upload.FileDescriptor)
			desc.Path = finalPath
			desc.Size = stat.Size()
			desc.Name = conf.Name
			desc.FileName = filepath.Base(finalPath)
			col = append(col, *desc)
		}
	}

	if len(col) != 0 {
		upload.UploadToS3(*instr.S3, col)
	}
}

func performExtraction(args []string) {
	if len(args) < 3 {
		log.Printf("Usage: %v ENCRYPTED_FILE KEY_FILE [OUT_DIRECTORY]\r\n", args[0])
		panicQuit()
	}

	target := args[1]
	keyfile := args[2]

	var out string

	var success bool
	if len(args) >= 4 {
		out = strings.Join([]string{args[3], "tmp"}, ".")
		success = core.DecryptFile(target, out, keyfile)
	} else {
		out = strings.Replace(target, ".enc", "", 1)
		out = strings.Join([]string{out, "tmp"}, ".")
		if out == target {
			log.Println("Cannot determine the out file, please specify")
			panicQuit()
		}
		success = core.DecryptFile(target, out, keyfile)
	}

	var ensureDirectory = regexp.MustCompile(`(\.(tmp|tgz|tar|gz))+`)
	directory := ensureDirectory.ReplaceAllString(out, "")

	if !core.PathExists(directory) {
		os.Mkdir(directory, 0770)
	}

	if success {
		core.ExtractArchive(out, directory)
	}

	core.CleanUpIfNeeded(out)

	if !success {
		panicQuit()
	}
}

func panicQuit() {
	os.Exit(1)
}
