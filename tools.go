package main

import "os"

func pathExists(path string) bool {
  _, checkErr := os.Stat(path)
  if (checkErr != nil && os.IsNotExist(checkErr)) {
    return false
  } else {
    return true
  }
}