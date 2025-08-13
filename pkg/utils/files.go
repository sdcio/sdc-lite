package utils

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"strings"
)

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false // Folder does not exist
	}
	return err == nil && !info.IsDir()
}

type FileWrapper struct {
	ref      string
	insecure bool
}

func NewFileWrapper(ref string) *FileWrapper {
	return &FileWrapper{
		ref: ref,
	}
}

func (f *FileWrapper) SetInsecure(b bool) {
	f.insecure = b
}

func (f *FileWrapper) ReadCloser() (io.ReadCloser, error) {
	switch {
	case f.ref == "-":
		return os.Stdin, nil
	case strings.HasPrefix(f.ref, "http"):

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: f.insecure},
		}
		client := &http.Client{Transport: tr}

		resp, err := client.Get(f.ref)
		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	default:
		file, err := os.Open(f.ref)
		if err != nil {
			return nil, err
		}
		return file, nil
	}
}

func (f *FileWrapper) Bytes() ([]byte, error) {
	rc, err := f.ReadCloser()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	return data, nil
}
