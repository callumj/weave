package app

import (
	"callumj.com/weave/core"
	"callumj.com/weave/remote"
	"callumj.com/weave/tools"
	"fmt"
	"io/ioutil"
	"os"
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
		core.ExtractArchive(out, directory)
	}

	tools.CleanUpIfNeeded(out)

	if len(eTag) != 0 {
		etagfile := fmt.Sprintf("%v/.weave.etag", directory)
		ioutil.WriteFile(etagfile, []byte(eTag), 0775)
	}

	if !success {
		panicQuit()
	}
}
