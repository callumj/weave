package core

import (
	"bytes"
	"fmt"
	"github.com/callumj/weave/remote/uptypes"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

type Configuration struct {
	Name      string
	Disabled  bool
	Except    []string
	ExceptReg *regexp.Regexp
	Only      []string
	OnlyReg   *regexp.Regexp
}

type Instruction struct {
	Src            string
	Encrypt        bool
	Configurations []Configuration
	Ignore         []string
	IgnoreReg      regexp.Regexp
	S3             *uptypes.S3Config
}

func ParseInstruction(path string) *Instruction {
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

	var ignoreReg = generateRegexpExpression(instr.Ignore)
	if ignoreReg == nil {
		log.Printf("Failed to merge into Regexp")
		return nil
	}

	instr.IgnoreReg = *ignoreReg

	for index, conf := range instr.Configurations {
		if !fillOutConfiguration(&conf) {
			log.Printf("Unable to compile except for %v\r\n", conf.Name)
		}
		instr.Configurations[index] = conf
	}

	return &instr
}

func fillOutConfiguration(conf *Configuration) bool {
	exceptLength := len(conf.Except)
	if exceptLength != 0 {
		exceptReg := generateRegexpExpression(conf.Except)
		if exceptReg == nil {
			log.Printf("Failed to merge %v except into Regexp", conf.Name)
			return false
		}
		conf.ExceptReg = exceptReg
	}

	onlyLength := len(conf.Only)
	if onlyLength != 0 {
		onlyReg := generateRegexpExpression(conf.Only)
		if onlyReg == nil {
			log.Printf("Failed to merge %v only into Regexp", conf.Name)
			return false
		}
		conf.OnlyReg = onlyReg
	}

	return true
}

func generateRegexpExpression(ary []string) *regexp.Regexp {
	var regBuffer bytes.Buffer
	total := len(ary)
	for index, ignorePat := range ary {
		regBuffer.WriteString(fmt.Sprintf("(%v)", ignorePat))
		if (index + 1) != total {
			regBuffer.WriteString("|")
		}
	}

	reg, err := regexp.Compile(regBuffer.String())
	if err != nil {
		return nil
	}

	return reg
}

func ExplainInstruction(instr Instruction) {
	log.Printf("Using: %v\r\n", instr.Src)
	log.Printf("Encryption: %v\r\n", instr.Encrypt)
	for _, conf := range instr.Configurations {
		exceptLength := len(conf.Except)
		if exceptLength != 0 {
			log.Printf("[%v] Except: %v", conf.Name, strings.Join(conf.Except, ", "))
		}
	}
}
