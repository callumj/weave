package main

import "os"
import "fmt"
import "path/filepath"

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
  fmt.Printf("Nase: %v", fullPath)

  // ensure working dir exists
  workingDir := fmt.Sprintf("%v/working", fullPath)
  _, checkErr := os.Stat(workingDir)
  if (checkErr != nil && os.IsNotExist(checkErr)) {
    mkErr := os.Mkdir(workingDir, 0775)
    if (mkErr != nil) {
      os.Exit(1)
    }
  }

  instr := parseInstruction(args[1])
  explainInstruction(instr)

  baseContents := getContents(instr.Src)
  suffix := fmt.Sprintf("%v/%v_%v.tar", workingDir, baseContents.Size, baseContents.Newest.Unix())
  createBaseArchive(instr.Src, baseContents.Contents, suffix)
}