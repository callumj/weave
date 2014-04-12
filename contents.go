package main

import (
  "os"
  "path/filepath"
  "log"
  "time"
)

type ContentsInfo struct {
  Size int64
  Newest time.Time
  Contents []string
}

func getContents(root string) ContentsInfo {
  cnts := ContentsInfo{}

  walkFn := func(path string, info os.FileInfo, err error) error {
    stat, err := os.Stat(path)
    if err != nil {
      return nil
    }

    if stat.Mode().IsRegular() {
      cnts.Size++
      if (cnts.Newest.Before(stat.ModTime())) {
        cnts.Newest = stat.ModTime()
      }
      cnts.Contents = append(cnts.Contents, path)
    }
    
    return nil
  }
  err := filepath.Walk(root, walkFn)
  if err != nil {
    log.Fatal(err)
  }

  return cnts
}