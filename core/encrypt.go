package core

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func DecryptFile(target, out, keyfile string) bool {
	keyContents, error := ioutil.ReadFile(keyfile)
	if error != nil {
		log.Printf("Could not load keyfile %v\r\n", keyfile)
		return false
	}

	inFile, err := os.Open(target)
	if err != nil {
		log.Printf("Could not load archive %v\r\n", target)
		return false
	}
	defer inFile.Close()

	block, err := aes.NewCipher(keyContents)
	if err != nil {
		log.Printf("Could not load AES Cipher %v\r\n", err)
		return false
	}

	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])

	outPntr, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0770)
	if err != nil {
		log.Printf("Could open %v for writing\r\n", out)
		return false
	}
	defer outPntr.Close()

	reader := &cipher.StreamReader{S: stream, R: inFile}
	// Copy the input file to the output file, decrypting as we go.
	if _, err := io.Copy(outPntr, reader); err != nil {
		panic(err)
	}

	return true
}

func EncryptFile(target, out, keyfile string) bool {
	keyContents, error := ioutil.ReadFile(keyfile)
	if error != nil {
		log.Printf("Could not load keyfile %v\r\n", keyfile)
		return false
	}

	inFile, err := os.Open(target)
	if err != nil {
		log.Printf("Could not load archive %v\r\n", target)
		return false
	}
	defer inFile.Close()

	block, err := aes.NewCipher(keyContents)
	if err != nil {
		log.Printf("Could not load AES Cipher %v\r\n", err)
		return false
	}

	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])

	outPntr, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0770)
	if err != nil {
		log.Printf("Could open %v for writing\r\n", out)
		return false
	}
	defer outPntr.Close()

	writer := &cipher.StreamWriter{S: stream, W: outPntr}

	if _, err := io.Copy(writer, inFile); err != nil {
		log.Printf("Could not encrypt\r\n", out)
		return false
	}
	writer.Close()
	return true
}
