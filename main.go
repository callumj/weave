package main

import (
  "fmt"
  "os"
  "path/filepath"
)

func main() {
  args := os.Args

  if len(args) == 1 {
    fmt.Printf("Usage: %v CONFIG_FILE", args[0]);
    fmt.Println()
    os.Exit(1);
  }

  abs, absErr := filepath.Abs(args[1])
  if (absErr != nil) {
    os.Exit(1)
  }
  fullPath := filepath.Dir(abs)

  // ensure working dir exists
  workingDir := fmt.Sprintf("%v/working", fullPath)
  if (!pathExists(workingDir)) {
    mkErr := os.Mkdir(workingDir, 0775)
    if (mkErr != nil) {
      os.Exit(1)
    }
  }

  instr := parseInstruction(args[1])
  explainInstruction(instr)

  baseContents := getContents(instr.Src)
  suffix := fmt.Sprintf("%v/%v_%v.tar", workingDir, baseContents.Size, baseContents.Newest.Unix())
  baseArchive := createBaseArchive(instr.Src, baseContents.Contents, suffix)

  for _, conf := range instr.Configurations {
    thisPath := fmt.Sprintf("%v/configurations/%v", fullPath, conf.Name)
    if (!pathExists(thisPath)) {
      fmt.Printf("%v does not exist", thisPath)
      fmt.Println()
    } else {
      thisContents := getContents(thisPath)
      tarPath := fmt.Sprintf("%v/%v_%v_%v.tar", workingDir, conf.Name, thisContents.Size, thisContents.Newest.Unix())
      mergeIntoBaseArchive(baseArchive, thisPath, thisContents.Contents, tarPath)
      gzipPath := fmt.Sprintf("%v.gz", tarPath)
      compressArchive(tarPath, gzipPath)
      os.Remove(tarPath)
    }
  }
}