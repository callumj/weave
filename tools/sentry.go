package tools

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type SentryStacktraceFrame struct {
	Function string `json:"function"`
}

type SentryStacktrace struct {
	Frames []SentryStacktraceFrame `json:"frames"`
}

type SentryPayload struct {
	Message string `json:"message"`

	EventID   string `json:"event_id"`
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Logger    string `json:"logger"`

	Platform   string                 `json:"platform,omitempty"`
	Culprit    string                 `json:"culprit,omitempty"`
	ServerName string                 `json:"server_name,omitempty"`
	Modules    []map[string]string    `json:"modules,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
	Tags       map[string]string      `json:"tags"`

	Stacktrace SentryStacktrace `json:"stacktrace"`
}

func HandleMessage(message, errorfile string) error {
	if len(LocalConfig.Sentry) == 0 {
		log.Println("Will not be sending message to Sentry, not configured")
		return nil
	}

	var projectID string
	uri, err := url.Parse(LocalConfig.Sentry)

	if err != nil {
		return err
	}

	if uri.User == nil {
		return errors.New("raven: dsn missing public key and/or private key")
	}
	publicKey := uri.User.Username()
	secretKey, ok := uri.User.Password()
	if !ok {
		return errors.New("raven: dsn missing private key")
	}
	uri.User = nil

	if idx := strings.LastIndex(uri.Path, "/"); idx != -1 {
		projectID = uri.Path[idx+1:]
		uri.Path = uri.Path[:idx+1] + "api/" + projectID + "/store/"
	}

	if projectID == "" {
		return errors.New("raven: dsn missing project id")
	}

	finalUrl := uri.String()
	authHeader := fmt.Sprintf("Sentry sentry_version=4, sentry_key=%s, sentry_secret=%s", publicKey, secretKey)

	pack := SentryPayload{
		Message:   message,
		EventID:   strconv.FormatInt(time.Now().Unix(), 10),
		Timestamp: time.Now().Format("2006-01-02T15:04:05"),
		Level:     "error",
		Logger:    "CLI",
		Tags:      make(map[string]string),
		Stacktrace: SentryStacktrace{
			Frames: []SentryStacktraceFrame{},
		},
	}

	hostname, err := os.Hostname()
	if err == nil {
		pack.Tags["hostname"] = hostname
	}

	if len(errorfile) != 0 {
		_, err = os.Stat(errorfile)
		if err != nil && os.IsNotExist(err) {
			return err
		}

		fp, err := os.Open(errorfile)
		if err != nil {
			return err
		}
		defer fp.Close()

		lines := []string{}
		scanner := bufio.NewScanner(fp)
		for scanner.Scan() {
			txt := scanner.Text()
			if len(txt) <= 5 {
				continue
			}

			if len(lines) > 20 {
				lines = lines[:len(lines)-1]
			}
			lines = append(lines, txt)
		}

		for index := len(lines) - 1; index >= 0; index-- {
			pack.Stacktrace.Frames = append(pack.Stacktrace.Frames, SentryStacktraceFrame{Function: lines[index]})
		}
	}

	jsonStr, err := json.Marshal(pack)
	if err != nil {
		return err
	}
	body := bytes.NewReader(jsonStr)
	req, _ := http.NewRequest("POST", finalUrl, body)
	req.Header.Set("X-Sentry-Auth", authHeader)
	req.Header.Set("User-Agent", "github.com/callumj/sentry-msg")
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%v\r\n", err)
		return err
	}

	respBody, err := ioutil.ReadAll(resp.Body)

	if err == nil {
		fmt.Printf("%v\r\n", string(respBody))
	}

	return nil
}
