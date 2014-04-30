package core

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type FileInfo struct {
	AbsPath    string
	RelPath    string
	ModTime    time.Time
	Size       int64
	Identifier string
}

type ContentsInfo struct {
	Size     int64
	Newest   time.Time
	Contents []FileInfo
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
			info := FileInfo{}
			info.ModTime = stat.ModTime()
			info.AbsPath = path
			info.RelPath = strings.Replace(path, root, "", 1)
			if strings.HasPrefix(info.RelPath, "/") {
				info.RelPath = strings.TrimPrefix(info.RelPath, "/")
			}
			info.Size = stat.Size()
			info.Identifier = generateIdentifier(info)
			cnts.Contents = append(cnts.Contents, info)
		}

		return nil
	}
	err := filepath.Walk(root, walkFn)
	if err != nil {
		log.Printf("Failed to determine listing of %v\r\n", root)
		return nil
	}

	sort.Sort(byPath(cnts.Contents))

	return &cnts
}

func FilterContents(existing ContentsInfo, ignore *regexp.Regexp, only *regexp.Regexp) *ContentsInfo {
	if ignore == nil && only == nil {
		return &existing
	}

	cnts := ContentsInfo{}

	for _, info := range existing.Contents {
		if ignore != nil && ignore.MatchString(info.RelPath) {
			continue
		}

		if only != nil && !only.MatchString(info.RelPath) {
			continue
		}

		cnts.Size++
		if cnts.Newest.Before(info.ModTime) {
			cnts.Newest = info.ModTime
		}

		cnts.Contents = append(cnts.Contents, info)
	}

	sort.Sort(byPath(cnts.Contents))

	return &cnts
}

func generateIdentifier(info FileInfo) string {
	str := fmt.Sprintf("%s_%v_%v", info.RelPath, info.ModTime.Unix(), info.Size)

	hasher := sha1.New()
	hasher.Write([]byte(str))
	hashed := hasher.Sum(nil)
	return hex.EncodeToString(hashed)
}

type byPath []FileInfo

func (v byPath) Len() int           { return len(v) }
func (v byPath) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v byPath) Less(i, j int) bool { return v[i].AbsPath < v[j].AbsPath }
