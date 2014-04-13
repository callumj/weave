package main

import (
  "fmt"
  "os"
  "path/filepath"
  "strings"
)

func main() {
  args := os.Args

  if len(args) == 1 {
    fmt.Printf("Usage: %v CONFIG_FILE", args[0]);
    fmt.Printf("Usage: %v ENCRYPTED_FILE OUT_FILE KEY_FILE", args[0]);
    fmt.Println()
    os.Exit(1);
  }

  if (strings.HasSuffix(args[1], ".enc")) {
    if len(args) != 4 {
      fmt.Printf("Usage: %v ENCRYPTED_FILE OUT_FILE KEY_FILE", args[0]);
      os.Exit(1);
    }

    target := args[1]
    out := args[2]
    keyfile := args[3]

    decryptFile(target, out, keyfile)
    return
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

      if (instr.Encrypt) {
        cryptPath := fmt.Sprintf("%v.enc", gzipPath)
        keyFile := fmt.Sprintf("%v/keys/%v", fullPath, conf.Name)
        if (!encryptFile(gzipPath, cryptPath, keyFile)) {
          fmt.Printf("Failed to encrypt %v\r\n", gzipPath)
        }
        os.Remove(gzipPath)
      }
    }
  }
}