package core

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

type ContentsInfo struct {
	Size     int64
	Newest   time.Time
	Contents []string
}

func GetContents(root string, ignoreReg regexp.Regexp) *ContentsInfo {
	cnts := ContentsInfo{}

	walkFn := func(path string, info os.FileInfo, err error) error {
		stat, err := os.Stat(path)
		if err != nil {
			return nil
		}

		if stat.Mode().IsRegular() {

			if ignoreReg.MatchString(path) {
				return nil
			}

			cnts.Size++
			if cnts.Newest.Before(stat.ModTime()) {
				cnts.Newest = stat.ModTime()
			}
			cnts.Contents = append(cnts.Contents, path)
		}

		return nil
	}
	err := filepath.Walk(root, walkFn)
	if err != nil {
		log.Printf("Failed to determine listing of %v\r\n", root)
		return nil
	}

	return &cnts
}
