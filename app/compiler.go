package app

import (
	"callumj.com/weave/core"
)

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

	var col []uptypes.FileDescriptor

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

			desc := new(uptypes.FileDescriptor)
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
