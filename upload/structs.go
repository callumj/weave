package upload

import "time"

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
