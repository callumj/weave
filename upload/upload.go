package upload

import (
	"callumj.com/weave/upload/s3"
	"callumj.com/weave/upload/uptypes"
)

func UploadToS3(config uptypes.S3Config, files []uptypes.FileDescriptor) {
	s3.UploadToS3(config, files)
}
