package main

import (
  "archive/tar"
  "log"
  "os"
  "io"
  "fmt"
  "strings"
)

func createArchive(basedir string, contents []string) {
  tarPntr, err := os.Create("/tmp/test.tar")
  if err != nil {
      log.Fatalln(err)
  }

  tw := tar.NewWriter(tarPntr)
  total := len(contents)

  for index, file := range contents {
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

    fmt.Printf("Completed %v / %v", index + 1, total)
    fmt.Println()
  }

  // Make sure to check the error on Close.
  if err := tw.Close(); err != nil {
    log.Fatalln(err)

  }

  tarPntr.Close()
}