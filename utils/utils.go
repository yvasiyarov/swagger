package utils

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

func StringSliceContains(stringSlice []string, data string) bool {
	for _, v := range stringSlice {
		if v == data {
			return true
		}
	}

	return false
}

// return gopath, goroot, err
func GetGoVars() (string, string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", "", errors.New("Please set the $GOPATH environment variable")
	}

	goroot := filepath.Clean(runtime.GOROOT())
	if goroot == "" {
		return "", "", errors.New("Please set $GOROOT environment variable")
	}

	return gopath, goroot, nil
}
