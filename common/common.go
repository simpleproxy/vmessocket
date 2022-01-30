package common

import (
	"fmt"
	"go/build"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vmessocket/vmessocket/common/errors"
)

//go:generate go run github.com/vmessocket/vmessocket/common/errors/errorgen

var ErrNoClue = errors.New("not enough information for making a decision")

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Must2(v interface{}, err error) interface{} {
	Must(err)
	return v
}

func Error2(v interface{}, err error) error {
	return err
}

func envFile() (string, error) {
	if file := os.Getenv("GOENV"); file != "" {
		if file == "off" {
			return "", fmt.Errorf("GOENV=off")
		}
		return file, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", fmt.Errorf("missing user-config dir")
	}
	return filepath.Join(dir, "go", "env"), nil
}

func GetRuntimeEnv(key string) (string, error) {
	file, err := envFile()
	if err != nil {
		return "", err
	}
	if file == "" {
		return "", fmt.Errorf("missing runtime env file")
	}
	var data []byte
	var runtimeEnv string
	data, readErr := os.ReadFile(file)
	if readErr != nil {
		return "", readErr
	}
	envStrings := strings.Split(string(data), "\n")
	for _, envItem := range envStrings {
		envItem = strings.TrimSuffix(envItem, "\r")
		envKeyValue := strings.Split(envItem, "=")
		if strings.EqualFold(strings.TrimSpace(envKeyValue[0]), key) {
			runtimeEnv = strings.TrimSpace(envKeyValue[1])
		}
	}
	return runtimeEnv, nil
}

func GetGOBIN() string {
	GOBIN := os.Getenv("GOBIN")
	if GOBIN == "" {
		var err error
		GOBIN, err = GetRuntimeEnv("GOBIN")
		if err != nil {
			return filepath.Join(build.Default.GOPATH, "bin")
		}
		if GOBIN == "" {
			return filepath.Join(build.Default.GOPATH, "bin")
		}
		return GOBIN
	}
	return GOBIN
}

func GetGOPATH() string {
	GOPATH := os.Getenv("GOPATH")
	if GOPATH == "" {
		var err error
		GOPATH, err = GetRuntimeEnv("GOPATH")
		if err != nil {
			return build.Default.GOPATH
		}
		if GOPATH == "" {
			return build.Default.GOPATH
		}
		return GOPATH
	}
	return GOPATH
}

func FetchHTTPContent(target string) ([]byte, error) {
	parsedTarget, err := url.Parse(target)
	if err != nil {
		return nil, newError("invalid URL: ", target).Base(err)
	}

	if s := strings.ToLower(parsedTarget.Scheme); s != "http" && s != "https" {
		return nil, newError("invalid scheme: ", parsedTarget.Scheme)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(&http.Request{
		Method: "GET",
		URL:    parsedTarget,
		Close:  true,
	})
	if err != nil {
		return nil, newError("failed to dial to ", target).Base(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, newError("unexpected HTTP status code: ", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, newError("failed to read HTTP response").Base(err)
	}

	return content, nil
}
