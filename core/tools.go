package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func GenerateNameSuffix(info ContentsInfo) string {
	var buffer bytes.Buffer
	for _, item := range info.Contents {
		buffer.WriteString(item.Identifier)
	}

	hasher := sha256.New()
	hasher.Write(buffer.Bytes())
	hashed := hasher.Sum(nil)
	return hex.EncodeToString(hashed)
}

func GenerateFinalNameSuffix(baseSuffix string, info ContentsInfo) string {
	hash := GenerateNameSuffix(info)
	join := fmt.Sprintf("%v_%v", hash, baseSuffix)
	hasher := sha256.New()
	hasher.Write([]byte(join))
	hashed := hasher.Sum(nil)
	return hex.EncodeToString(hashed)
}
