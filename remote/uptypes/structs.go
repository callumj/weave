package uptypes

import "time"

type FileDescriptor struct {
	Name     string
	FileName string
	Path     string
	Size     int64
}

type Content struct {
	Key          string
	LastModified time.Time
	ETag         string
	Size         int64
	StorageClass string
}

type ListBucketResult struct {
	Name        string
	Prefix      string
	Marker      string
	MaxKeys     int
	IsTruncated bool
	Contents    []Content
}

type S3Config struct {
	Bucket     string
	Access_Key string
	Secret     string
	Folder     string
	Endpoint   string
	Public     bool
}
