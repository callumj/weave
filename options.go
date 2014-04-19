package main

import (
	"bytes"
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
	IgnoreReg      regexp.Regexp
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

	var regBuffer bytes.Buffer
	total := len(instr.Ignore)
	for index, ignorePat := range instr.Ignore {
		regBuffer.WriteString(fmt.Sprintf("(%v)", ignorePat))
		if (index + 1) != total {
			regBuffer.WriteString("|")
		}
	}

	reg, err := regexp.Compile(regBuffer.String())
	if err != nil {
		log.Printf("Failed to merge into Regexp")
		return nil
	}
	instr.IgnoreReg = *reg

	return &instr
}

func explainInstruction(instr Instruction) {
	fmt.Printf("Using: %v\r\n", instr.Src)
	fmt.Printf("Encryption: %v\r\n", instr.Encrypt)
}
