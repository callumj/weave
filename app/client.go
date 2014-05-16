package app

import (
	"fmt"
	"github.com/callumj/weave/core"
	"github.com/callumj/weave/remote"
	"github.com/callumj/weave/tools"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

	directory, _ = filepath.Abs(directory)

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
		runCallback("post_extraction", directory)
	}

	if !success {
		panicQuit()
	}
}

func runCallback(callback, directory string) {
	callbackPath := fmt.Sprintf("%v/%v.sh", directory, callback)
	if tools.PathExists(callbackPath) {
		bashLoc, _ := exec.LookPath("bash")
		if len(bashLoc) == 0 {
			bashLoc = "/bin/bash"
		}
		cmd := exec.Command(bashLoc, callbackPath)
		cmd.Dir = directory
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			panicQuitf("Unable to open STDOUT")
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			panicQuitf("Unable to open STDERR")
		}

		if err := cmd.Start(); err != nil {
			panicQuitf("Unable to start %v (%v)\r\n", callbackPath, err)
		}

		outLogPath := fmt.Sprintf("%v/%v.stdout.log", directory, callback)
		stdoutLogFile, err := os.Create(outLogPath)
		if err != nil {
			panicQuitf("Unable to open file for STDOUT logging %v\r\n", outLogPath)
		}
		defer stdoutLogFile.Close()

		errLogPath := fmt.Sprintf("%v/%v.stderr.log", directory, callback)
		stderrLogFile, err := os.Create(errLogPath)
		if err != nil {
			panicQuitf("Unable to open file for STDERR logging %v\r\n", errLogPath)
		}
		defer stderrLogFile.Close()

		go io.Copy(stdoutLogFile, stdout)
		go io.Copy(stderrLogFile, stderr)

		if err := cmd.Wait(); err != nil {
			panicQuitf("Failed finalise command (%v)\r\n", err)
		}
	}
}
