package app

import (
	"callumj.com/weave/core"
	"callumj.com/weave/remote"
	"callumj.com/weave/tools"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func performExtraction(args []string) {
	if len(args) < 3 {
		panicQuitf("Usage: %v ENCRYPTED_FILE KEY_FILE [OUT_DIRECTORY]\r\n", args[0])
	}

	target := args[1]
	keyfile := args[2]

	var eTag string

	matched, _ := regexp.MatchString("^https?:\\/\\/", target)
	if matched {
		if len(args) == 3 {
			panicQuitf("A OUT_DIRECTORY must be specified if using HTTP\r\n")
		} else {
			resp := remote.DownloadRemoteFile(target, args[3])
			if resp == nil {
				panicQuit()
			} else {
				target = resp.FilePath
				eTag = resp.ETag
			}
		}
	}

	var out string

	var success bool
	if len(args) >= 4 {
		out = strings.Join([]string{args[3], "tmp"}, ".")
		success = core.DecryptFile(target, out, keyfile)
	} else {
		out = strings.Replace(target, ".enc", "", 1)
		out = strings.Join([]string{out, "tmp"}, ".")
		if out == target {
			panicQuitf("Cannot determine the out file, please specify")
		}
		success = core.DecryptFile(target, out, keyfile)
	}

	var ensureDirectory = regexp.MustCompile(`(\.(tmp|tgz|tar|gz))+`)
	directory := ensureDirectory.ReplaceAllString(out, "")

	if !tools.PathExists(directory) {
		os.Mkdir(directory, 0770)
	}

	if success {
		success = core.ExtractArchive(out, directory)
	}

	tools.CleanUpIfNeeded(out)

	if success && len(eTag) != 0 {
		etagfile := fmt.Sprintf("%v/.weave.etag", directory)
		ioutil.WriteFile(etagfile, []byte(eTag), 0775)
	}

	if success {
		runPostExtractionCallback(directory)
	}

	if !success {
		panicQuit()
	}
}

func runPostExtractionCallback(directory string) {
	postExtractionPath := fmt.Sprintf("%v/post_extraction.sh", directory)
	if tools.PathExists(postExtractionPath) {
		cmd := exec.Command("/bin/sh", postExtractionPath)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			panicQuitf("Unable to open STDOUT")
		}
		if err := cmd.Start(); err != nil {
			panicQuitf("Unable to start %v (%v)\r\n", postExtractionPath, err)
		}

		logPath := fmt.Sprintf("%v/post_extraction.log", directory)
		stdoutLogFile, err := os.Create(logPath)
		if err != nil {
			panicQuitf("Unable to open file for logging %v\r\n", logPath)
		}
		defer stdoutLogFile.Close()

		_, err = io.Copy(stdoutLogFile, stdout)
		if err != nil {
			panicQuitf("Failed to copy STDOUT to file\r\n")
		}

		if err := cmd.Wait(); err != nil {
			panicQuitf("Failed finalise command (%v)\r\n", err)
		}
	}
}
