package s3

import (
	"callumj.com/weave/remote/uptypes"
	"fmt"
	"log"
	"strings"
)

func UploadToS3(config uptypes.S3Config, files []uptypes.FileDescriptor) {
	wrapped := buildWrappedConfig(config)

	for _, wr := range getFilesRequiringUpload(*wrapped, files) {
		file := wr.File
		log.Printf("Updating %v\r\n", file.Name)
		var suffix string
		if strings.Contains(file.Path, ".enc") {
			suffix = "tar.gz.enc"
		} else {
			suffix = "tar.gz"
		}
		filename := fmt.Sprintf("%s.%s", file.Name, suffix)
		itemUrl := fmt.Sprintf("%s/%s", wrapped.PuttableAddress, filename)
		putFile(file, itemUrl, wrapped.Keys, wr.AlreadyExists)
	}
}
