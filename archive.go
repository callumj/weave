package main

import (
  "archive/tar"
  "log"
  "os"
  "io"
  "fmt"
  "strings"
  "compress/gzip"
)

type Item struct {
  Start int64
  Length int64
  Name string
}

type ArchiveInfo struct {
  Items []Item
  Path string
}

func compressArchive(archivePath, outPath string) {
  dupe, dupeErr := os.Create(outPath)
  if (dupeErr != nil) {
    panic(dupeErr)
  }
  gzipPntr := gzip.NewWriter(dupe)

  basePntr, baseErr := os.Open(archivePath)
  if (baseErr != nil) {
    panic(baseErr)
  }

  io.Copy(gzipPntr, basePntr)

  gzipPntr.Close()

  basePntr.Close()
  dupe.Close()
}

func mergeIntoBaseArchive(baseArchive ArchiveInfo, basedir string, contents []string, file string) {
  // tar pntr for copy
  dupe, dupeErr := os.Create(file)
  if (dupeErr != nil) {
    panic(dupeErr)
  }
  tw := tar.NewWriter(dupe)

  basePntr, baseErr := os.Open(baseArchive.Path)
  if (baseErr != nil) {
    panic(baseErr)
  }

  io.Copy(dupe, basePntr)

  // bump to the end
  dupe.Seek(-2<<9, os.SEEK_END)

  // insert
  for _, item := range contents {
    writeFileToArchive(dupe, tw, item, basedir)
  }

  tw.Close()

  basePntr.Close()
  dupe.Close()
}

func createBaseArchive(basedir string, contents []string, file string) ArchiveInfo {
  tarPntr, err := os.Create(file)
  if err != nil {
      log.Fatalln(err)
  }

  tw := tar.NewWriter(tarPntr)
  total := len(contents)

  a := ArchiveInfo{Path: file}

  for index, file := range contents {
    item := writeFileToArchive(tarPntr, tw, file, basedir)
    fmt.Printf("\rArchiving %v / %v", index + 1, total)
    a.Items = append(a.Items, item)
  }

  // Make sure to check the error on Close.
  if err := tw.Close(); err != nil {
    log.Fatalln(err)

  }

  tarPntr.Close()
  fmt.Println()

  return a
}

func writeFileToArchive(tarPntr *os.File, tw *tar.Writer, file string, basedir string) Item {
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

  return Item{Start: curPos, Length: (endPos - curPos), Name: hdr.Name}
}