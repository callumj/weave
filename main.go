package main

import "os"
import "fmt"

func main() {
  args := os.Args

  if len(args) == 1 {
    fmt.Printf("Usage: %v DIRECTORY", args[0]);
    fmt.Println()
    os.Exit(1);
  }

  files := getContents(args[1])
  createArchive(args[1], files)
}