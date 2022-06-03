package goVods

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/grafov/m3u8"
)

type HlsDl struct {
	client        *http.Client
	directoryPath string
	hlsDpi        DomainPathIdentifier
}

func NewHlsDl(hlsUrl DomainPathIdentifier, dir string) *HlsDl {
	client := &http.Client{}
	return &HlsDl{client, dir, hlsUrl}
}

func newRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (hlsDl *HlsDl) DownloadSegment(dst io.Writer, segment *m3u8.MediaSegment) error {
	url := hlsDl.hlsDpi.GetSegmentChunkedUrl(segment)
	req, err := newRequest(url)
	if err != nil {
		return err
	}
	res, err := hlsDl.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return errors.New(res.Status)
	}
	_, err = io.Copy(dst, res.Body)
	return err
}

func (hlsDl *HlsDl) DownloadSegments(segments []*m3u8.MediaSegment) error {
	filePath := filepath.Join(hlsDl.directoryPath, "index.ts")
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()
	for _, segment := range segments {
		fmt.Println("Downloading", segment.URI)
		err = hlsDl.DownloadSegment(out, segment)
		if err != nil {
			return err
		}
	}
	return nil
}
