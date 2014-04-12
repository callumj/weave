package main

import (
  "archive/tar"
  "log"
  "os"
  "io"
  "fmt"
  "strings"
)

type Item struct {
  Start int64
  Length int64
  Name string
}

type ArchiveInfo struct {
  Items []Item
}

func createBaseArchive(basedir string, contents []string, file string) {
  tarPntr, err := os.Create(file)
  if err != nil {
      log.Fatalln(err)
  }

  tw := tar.NewWriter(tarPntr)
  total := len(contents)

  a := ArchiveInfo{}

  for index, file := range contents {
    curPos, posErr := tarPntr.Seek(0, 1)
    if (posErr != nil) {
      log.Fatalln(posErr)
    }
    stat, statErr := os.Stat(file)
    if (statErr != nil) {
      log.Fatalln(statErr)
    }

    hdr := &tar.Header{
      Name: strings.Replace(file, basedir, "", 1),
      Size: stat.Size(),
      Mode: 775,
      ModTime: stat.ModTime(),
    }
    if err := tw.WriteHeader(hdr); err != nil {
      log.Fatalln(err)
    }

    filePntr, fileErr := os.Open(file)
    if (fileErr != nil) {
      log.Fatalln(fileErr)
    }

    // read in chunks for memory
    buf := make([]byte, 1024)
    for {
      // read a chunk
      n, readErr := filePntr.Read(buf)
      if readErr != nil && readErr != io.EOF { panic(readErr) }
      if n == 0 { break }

      // write a chunk
      if _, tarWriteErr := tw.Write(buf[:n]); tarWriteErr != nil {
        log.Fatalln(tarWriteErr)
      }
    }

    if fileCloseErr := filePntr.Close(); fileCloseErr != nil {
      panic(fileCloseErr)
    }

    endPos, endPosErr := tarPntr.Seek(0, 1)
    if (endPosErr != nil) {
      log.Fatalln(endPosErr)
    }

    info := Item{Start: curPos, Length: (endPos - curPos), Name: hdr.Name}
    a.Items = append(a.Items, info)

    fmt.Printf("Completed %v / %v", index + 1, total)
    fmt.Println()
  }

  // Make sure to check the error on Close.
  if err := tw.Close(); err != nil {
    log.Fatalln(err)

  }

  tarPntr.Close()
}