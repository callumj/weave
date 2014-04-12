package main

import (
  "os"
  "path/filepath"
  "log"
)

func getContents(root string) []string {
  var store []string

  walkFn := func(path string, info os.FileInfo, err error) error {
    stat, err := os.Stat(path)
    if err != nil {
      return nil
    }

    if !stat.IsDir() {
      store = append(store, path)
    }
    
    return nil
  }
  err := filepath.Walk(root, walkFn)
  if err != nil {
    log.Fatal(err)
  }

  return store
}