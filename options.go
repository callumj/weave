package main

import (
	"fmt"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"log"
	"regexp"
)

type Configuration struct {
	Name     string
	Disabled bool
}

type Instruction struct {
	Src            string
	Encrypt        bool
	Configurations []Configuration
	Ignore         []string
	IgnoreReg      []regexp.Regexp
}

func parseInstruction(path string) *Instruction {
	fileContents, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Unable to find configuration %v\r\n", path)
		return nil
	}
	instr := Instruction{}

	err = yaml.Unmarshal([]byte(fileContents), &instr)
	if err != nil {
		log.Printf("Unable to parse configuration %v\r\n", path)
		return nil
	}

	for _, ignorePat := range instr.Ignore {
		reg, err := regexp.Compile(ignorePat)
		if err != nil {
			log.Printf("Unable to compile %v to Regex\r\n", ignorePat)
			return nil
		}
		instr.IgnoreReg = append(instr.IgnoreReg, *reg)
	}

	return &instr
}

func explainInstruction(instr Instruction) {
	fmt.Printf("Using: %v\r\n", instr.Src)
	fmt.Printf("Encryption: %v\r\n", instr.Encrypt)
}
