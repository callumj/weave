package upload

import (
	"encoding/xml"
	"fmt"
	"github.com/kr/s3"
	"github.com/kr/s3/s3util"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type S3Config struct {
	Bucket     string
	Access_Key string
	Secret     string
	Folder     string
	Endpoint   string
}

type FileDescriptor struct {
	Name     string
	FileName string
	Path     string
	Size     int64
}

type wrappedS3Details struct {
	Config          S3Config
	Endpoint        string
	Keys            s3.Keys
	PuttableAddress string
}

type wrappedStateInfo struct {
	File          FileDescriptor
	AlreadyExists bool
}

type wrappedSizeComp struct {
	Key  string
	Size int64
}

func UploadToS3(config S3Config, files []FileDescriptor) {
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

func getFilesRequiringUpload(wrapped wrappedS3Details, files []FileDescriptor) []wrappedStateInfo {
	fileMap := make(map[string]FileDescriptor)
	for _, file := range files {
		fileMap[file.Name] = file
	}

	reg, err := regexp.Compile("\\.tar.gz(.enc)?$")
	if err != nil {
		log.Printf("Failed to compile Regexp\r\n")
		return []wrappedStateInfo{}
	}

	allKnownFiles := getExistingFiles(wrapped)

	sizeMap := make(map[string]wrappedSizeComp)
	for _, bucketItem := range allKnownFiles {
		name := reg.ReplaceAllString(bucketItem.Key, "")
		if len(wrapped.Config.Folder) != 0 {
			name = strings.Replace(name, fmt.Sprintf("%v/", wrapped.Config.Folder), "", 1)
		}

		info := wrappedSizeComp{Key: bucketItem.Key, Size: bucketItem.Size}
		sizeMap[name] = info
	}

	requiringUpload := []wrappedStateInfo{}
	keysRequiringDeepLook := make(map[string]FileDescriptor)

	for name, file := range fileMap {
		if sizeInfo, found := sizeMap[name]; found {
			if sizeInfo.Size == file.Size {
				keysRequiringDeepLook[sizeInfo.Key] = file
			} else {
				info := wrappedStateInfo{File: file, AlreadyExists: true}
				requiringUpload = append(requiringUpload, info)
			}
		} else {
			info := wrappedStateInfo{File: file, AlreadyExists: false}
			requiringUpload = append(requiringUpload, info)
		}
	}

	for bucketKey, file := range keysRequiringDeepLook {
		name := getBucketItemProperName(wrapped, bucketKey)
		if name == "-1" {
			continue
		}
		if name != file.FileName {
			info := wrappedStateInfo{File: file, AlreadyExists: true}
			requiringUpload = append(requiringUpload, info)
		}
	}

	return requiringUpload
}

func getBucketItemProperName(wrapped wrappedS3Details, key string) string {
	reqUrl := fmt.Sprintf("%v%v", wrapped.Endpoint, key)
	r, _ := http.NewRequest("HEAD", reqUrl, nil)
	r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	s3.Sign(r, wrapped.Keys)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Printf("Error requesting %v\r\n", reqUrl)
		return "-1"
	}
	return resp.Header.Get("x-amz-meta-fullname")
}

func buildWrappedConfig(config S3Config) *wrappedS3Details {
	keys := new(s3.Keys)
	keys.AccessKey = config.Access_Key
	keys.SecretKey = config.Secret

	var endpoint string
	if len(config.Endpoint) != 0 {
		endpoint = fmt.Sprintf("https://s3-%v.amazonaws.com", config.Endpoint)
	} else {
		endpoint = "https://s3.amazonaws.com"
	}

	wrapped := new(wrappedS3Details)
	wrapped.Config = config
	wrapped.Endpoint = fmt.Sprintf("%v/%v/", endpoint, config.Bucket)
	wrapped.Keys = *keys

	if len(config.Folder) == 0 {
		wrapped.PuttableAddress = wrapped.Endpoint
	} else {
		wrapped.PuttableAddress = fmt.Sprintf("%v%v", wrapped.Endpoint, config.Folder)
	}

	return wrapped
}

func putFile(desc FileDescriptor, restUrl string, keys s3.Keys, alreadyExists bool) bool {
	if alreadyExists {
		deleteBucketItem(restUrl, keys)
	}

	fr, err := os.Open(desc.Path)
	if err != nil {
		log.Printf("Failed to open %v\r\n", desc.Path)
		return false
	}
	defer fr.Close()

	conf := new(s3util.Config)
	conf.Keys = &keys
	conf.Service = s3util.DefaultConfig.Service

	h := http.Header{
		"x-amz-meta-fullname": {desc.FileName},
	}

	s3Wr, err := s3util.Create(restUrl, h, conf)
	if err != nil {
		log.Printf("Failed to open object for writing %v %v\r\n", restUrl, err)
		return false
	}
	defer s3Wr.Close()
	io.Copy(s3Wr, fr)

	return true
}

func getExistingFiles(details wrappedS3Details) []Content {
	var finalUrl string
	if len(details.Config.Folder) == 0 {
		finalUrl = details.Endpoint
	} else {
		finalUrl = fmt.Sprintf("%v?prefix=%v/", details.Endpoint, details.Config.Folder)
	}

	allResults := getBucketContents(finalUrl, details.Keys)

	contents := []Content{}
	for _, result := range allResults {
		contents = append(contents, result.Contents...)
	}

	return contents
}

func deleteBucketItem(restUrl string, keys s3.Keys) bool {
	r, _ := http.NewRequest("DELETE", restUrl, nil)
	r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	s3.Sign(r, keys)

	_, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Printf("Failed to delete %v %v\r\n", restUrl, err)
		return false
	}

	return true
}

func getBucketContents(restUrl string, keys s3.Keys) []ListBucketResult {
	r, _ := http.NewRequest("GET", restUrl, nil)
	r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	s3.Sign(r, keys)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	res := ListBucketResult{}
	err = xml.Unmarshal([]byte(body), &res)

	if res.IsTruncated {
		lastItem := res.Contents[len(res.Contents)-1]
		nextUrl, err := url.Parse(restUrl)
		if err != nil {
			return []ListBucketResult{}
		}
		q := nextUrl.Query()
		q.Set("marker", lastItem.Key)
		nextUrl.RawQuery = q.Encode()

		nextUrlString := fmt.Sprintf("%s", nextUrl)
		nextBucketContents := getBucketContents(nextUrlString, keys)

		finalResult := []ListBucketResult{res}
		return append(finalResult, nextBucketContents...)
	} else {
		return []ListBucketResult{res}
	}
}
