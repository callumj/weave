package app

import (
	"fmt"
	"github.com/callumj/weave/core"
	"github.com/callumj/weave/remote"
	"github.com/callumj/weave/remote/uptypes"
	"github.com/callumj/weave/tools"
	"log"
	"os"
	"path/filepath"
)

func performCompilation(configPath string, options option_set) {
	fullPath := filepath.Dir(configPath)

	// ensure working dir exists
	workingDir := fmt.Sprintf("%v/working", fullPath)
	if !tools.PathExists(workingDir) {
		log.Println("Working directory does not existing, creating")
		err := os.Mkdir(workingDir, 0775)
		if err != nil {
			panicQuitf("Unable to create %v\r\n", workingDir)
		}
	}

	instr := core.ParseInstruction(configPath)
	if instr == nil {
		panicQuit()
	}
	core.ExplainInstruction(instr)

	baseContents := core.GetContents(instr.Src, &instr.IgnoreReg)
	if baseContents == nil {
		panicQuit()
	}

	baseSuffix := core.GenerateNameSuffix(*baseContents)
	baseFileName := fmt.Sprintf("%v/%v.tar", workingDir, baseSuffix)
	baseArchive := core.CreateBaseArchive(instr.Src, baseContents.Contents, baseFileName)

	if baseArchive == nil {
		panicQuitf("Failed to create base archive.")
	}

	var col []uptypes.FileDescriptor

	for _, conf := range instr.Configurations {
		if len(options.OnlyRun) != 0 {
			if options.OnlyRun != conf.Name {
				continue
			}
		}
		finalPath := processConfiguration(conf, fullPath, instr, baseContents, baseArchive)
		if instr.S3 != nil && options.DisableS3 != true {
			col = appendForS3(finalPath, conf, col)
		}
	}

	if len(col) != 0 {
		remote.UploadToS3(*instr.S3, col)
	}
}

func processConfiguration(conf core.Configuration, fullPath string, instr *core.Instruction, baseContents *core.ContentsInfo, baseArchive *core.ArchiveInfo) string {
	thisPath := fmt.Sprintf("%v/configurations/%v", fullPath, conf.Name)
	workingDir := fmt.Sprintf("%v/working", fullPath)
	log.Printf("Configuring: %v\r\n", thisPath)

	thisContents := constructContents(thisPath, baseContents, instr)

	filteredContents := core.FilterContents(*baseContents, conf.ExceptReg, conf.OnlyReg)
	recalcBaseSuffix := core.GenerateNameSuffix(*filteredContents)
	tarPath := fmt.Sprintf("%v/%v_%v.tar", workingDir, conf.Name, core.GenerateFinalNameSuffix(recalcBaseSuffix, *thisContents))

	if !core.MergeIntoBaseArchive(*baseArchive, thisPath, thisContents.Contents, tarPath, filteredContents) {
		panicQuitf("Failed to merge with base archive. Quitting.")
	}
	gzipPath := fmt.Sprintf("%v.gz", tarPath)
	core.CompressArchive(tarPath, gzipPath)
	os.Remove(tarPath)

	finalPath := gzipPath
	if instr.Encrypt {
		cryptPath := fmt.Sprintf("%v.enc", gzipPath)
		keyFile := fmt.Sprintf("%v/keys/%v", fullPath, conf.Name)
		if !core.EncryptFile(gzipPath, cryptPath, keyFile) {
			panicQuitf("Failed to encrypt %v. Quiting..\r\n", gzipPath)
		} else {
			finalPath = cryptPath
		}
		os.Remove(gzipPath)
	}

	return finalPath
}

func appendForS3(finalPath string, conf core.Configuration, col []uptypes.FileDescriptor) []uptypes.FileDescriptor {
	stat, err := os.Stat(finalPath)
	if err != nil {
		panicQuitf("Unable to query %v\r\n", finalPath)
	}

	desc := new(uptypes.FileDescriptor)
	desc.Path = finalPath
	desc.Size = stat.Size()
	desc.Name = conf.Name
	desc.FileName = filepath.Base(finalPath)
	return append(col, *desc)
}

func constructContents(thisPath string, baseContents *core.ContentsInfo, instr *core.Instruction) *core.ContentsInfo {
	var thisContents *core.ContentsInfo
	if tools.PathExists(thisPath) {
		thisContents = core.GetContents(thisPath, &instr.IgnoreReg)
	} else {
		thisContents = new(core.ContentsInfo)
		thisContents.Size = 0
		thisContents.Contents = []core.FileInfo{}
		thisContents.Newest = baseContents.Newest
	}

	return thisContents
}
