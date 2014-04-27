package upload

import (
	"fmt"
	"github.com/kr/s3"
	"io/ioutil"
	"log"
	"net/http"
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
	Path string
	Size int64
}

func UploadToS3(config S3Config, files []FileDescriptor) {
	keys := new(s3.Keys)
	keys.AccessKey = config.Access_Key
	keys.SecretKey = config.Secret

	var endpoint string
	if len(config.Endpoint) != 0 {
		endpoint = fmt.Sprintf("https://s3-%v.amazonaws.com", config.Endpoint)
	} else {
		endpoint = "https://s3.amazonaws.com"
	}

	url := fmt.Sprintf("%v/%v/", endpoint, config.Bucket)
	r, _ := http.NewRequest("GET", url, nil)
	r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	s3.Sign(r, *keys)

	fmt.Printf("%v\r\n", r.Header)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("%s\r\n", body)
}
