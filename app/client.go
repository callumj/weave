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
	if len(args) < 2 {
		panicQuitf("Usage: %v ENCRYPTED_FILE KEY_FILE [OUT_DIRECTORY]\r\n", args[0])
	}

	target := args[0]
	keyfile := args[1]

	var eTag string

	matched, _ := regexp.MatchString("^https?:\\/\\/", target)
	if matched {
		if len(args) == 2 {
			panicQuitf("A OUT_DIRECTORY must be specified if using HTTP\r\n")
		} else {
			resp := remote.DownloadRemoteFile(target, args[2])
			if resp == nil {
				panicQuit()
			} else {
				target = resp.FilePath
				eTag = resp.ETag
			}
		}
	}
	defer os.Remove(target)

	var out string

	var success bool
	if len(args) >= 3 {
		out = strings.Join([]string{args[2], "tmp"}, ".")
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

	preExtraction := core.FetchFile(out, "/pre_extraction.sh")
	if len(preExtraction) != 0 {
		// write the pre file
		preName := fmt.Sprintf("%v/pre_extraction.sh", directory)
		ioutil.WriteFile(preName, []byte(preExtraction), 0770)
		runCallback(preName, directory)
	}

	if success {
		success = core.ExtractArchive(out, directory)
	}

	tools.InitConfig(fmt.Sprintf("%s/local_config.yml", directory))
	tools.CleanUpIfNeeded(out)

	if success && len(eTag) != 0 {
		etagfile := fmt.Sprintf("%v/.weave.etag", directory)
		ioutil.WriteFile(etagfile, []byte(eTag), 0775)
	}

	if success {
		callbackPath := fmt.Sprintf("%v/post_extraction.sh", directory)
		runCallback(callbackPath, directory)
	}

	if !success {
		panicQuit()
	}
}

func runCallback(callbackPath, directory string) {
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

		outLogPath := fmt.Sprintf("%v.stdout.log", callbackPath)
		stdoutLogFile, err := os.Create(outLogPath)
		if err != nil {
			panicQuitf("Unable to open file for STDOUT logging %v\r\n", outLogPath)
		}
		defer stdoutLogFile.Close()

		errLogPath := fmt.Sprintf("%v.stderr.log", callbackPath)
		stderrLogFile, err := os.Create(errLogPath)
		if err != nil {
			panicQuitf("Unable to open file for STDERR logging %v\r\n", errLogPath)
		}
		defer stderrLogFile.Close()

		go io.Copy(stdoutLogFile, stdout)
		go io.Copy(stderrLogFile, stderr)

		if err := cmd.Wait(); err != nil {
			outStat, outErr := stdoutLogFile.Stat()
			if outErr == nil && outStat.Size() > 0 {
				tools.HandleMessage(fmt.Sprintf("STDOUT: Error with %s", callbackPath), outLogPath)
			}

			errStat, outErr := stdoutLogFile.Stat()
			if outErr == nil && errStat.Size() > 0 {
				tools.HandleMessage(fmt.Sprintf("STDERR: Error with %s", callbackPath), errLogPath)
			}
			panicQuitf("Failed finalise command (%v)\r\n", err)
		}
	}
}
