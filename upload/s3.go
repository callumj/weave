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
	Name string
	Path string
	Size int64
}

type wrappedS3Details struct {
	Config          S3Config
	Endpoint        string
	Keys            s3.Keys
	PuttableAddress string
}

func UploadToS3(config S3Config, files []FileDescriptor) {
	wrapped := buildWrappedConfig(config)

	getFilesRequiringUpload(*wrapped, files)
}

func getFilesRequiringUpload(wrapped wrappedS3Details, files []FileDescriptor) {
	sizeMap := make(map[string]int64)
	for _, file := range files {
		sizeMap[file.Name] = file.Size
	}

	reg, err := regexp.Compile("\\.tar.gz(.enc)?$")
	if err != nil {
		log.Printf("Failed to compile Regexp\r\n")
		return
	}

	keysRequiringDeepLook := []string{}
	allKnownFiles := getExistingFiles(wrapped)
	for _, bucketItem := range allKnownFiles {
		name := reg.ReplaceAllString(bucketItem.Key, "")
		if len(wrapped.Config.Folder) != 0 {
			name = strings.Replace(name, fmt.Sprintf("%v/", wrapped.Config.Folder), "", 1)
		}
		if val, found := sizeMap[name]; found {
			if val == bucketItem.Size {
				keysRequiringDeepLook = append(keysRequiringDeepLook, bucketItem.Key)
			}
		}
	}
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

func putFiles(wrapped wrappedS3Details, files []FileDescriptor) {
	for _, item := range files {
		var suffix string
		if strings.Contains(item.Path, ".enc") {
			suffix = "tar.gz.enc"
		} else {
			suffix = "tar.gz"
		}
		filename := fmt.Sprintf("%s.%s", item.Name, suffix)
		itemUrl := fmt.Sprintf("%s/%s", wrapped.PuttableAddress, filename)
		putFile(item, itemUrl, wrapped.Keys)
	}
}

func putFile(desc FileDescriptor, restUrl string, keys s3.Keys) bool {
	fr, err := os.Open(desc.Path)
	if err != nil {
		log.Printf("Failed to open %v\r\n", desc.Path)
		return false
	}
	defer fr.Close()

	conf := new(s3util.Config)
	conf.Keys = &keys
	conf.Service = s3util.DefaultConfig.Service

	s3Wr, err := s3util.Create(restUrl, nil, conf)
	if err != nil {
		log.Printf("Failed to open object for writing %v\r\n", restUrl)
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
