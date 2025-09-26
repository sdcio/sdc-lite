package utils

import (
	"context"
	"crypto/tls"
	"fmt"
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

func (f *FileWrapper) ReadCloser(ctx context.Context) (io.ReadCloser, error) {
	switch {
	case f.ref == "" || f.ref == "-":
		return NewCtxReader(ctx, os.Stdin), nil
	case strings.HasPrefix(f.ref, "http"):

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: f.insecure},
		}
		client := &http.Client{Transport: tr}

		req, err := http.NewRequestWithContext(ctx, "GET", f.ref, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			resp.Body.Close()
			return nil, fmt.Errorf("http request for %s failed with: %s", f.ref, resp.Status)
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
	rc, err := f.ReadCloser(context.Background())
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
