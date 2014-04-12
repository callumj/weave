package main

import (
  "gopkg.in/yaml.v1"
  "io/ioutil"
  "log"
  "fmt"
)

type Configuration struct {
  Name string
  Disabled bool 
}

type Instruction struct {
  Src string
  Encrypt bool
  Configurations []Configuration
}

func parseInstruction(path string) Instruction {
  fileContents, fileErr := ioutil.ReadFile(path)
  if (fileErr != nil) {
    log.Fatal(fileErr)
  }
  instr := Instruction{}

  ymlErr := yaml.Unmarshal([]byte(fileContents), &instr)
  if ymlErr != nil {
    log.Fatal(ymlErr)
  }

  return instr
}

func explainInstruction(instr Instruction) {
  fmt.Printf("Using: %v\r\n", instr.Src)
  fmt.Printf("Encryption: %v\r\n", instr.Encrypt)
}