package goVods

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/grafov/m3u8"
	"github.com/samber/lo"
)

type VideoData struct {
	StreamerName string
	VideoId      string
	Time         time.Time
}

func (videoData *VideoData) GetUrlPathUniqueIdentifier() string {
	unixTime := videoData.Time.Unix()
	baseUrl := videoData.StreamerName + "_" + videoData.VideoId + "_" + fmt.Sprint(unixTime)
	hasher := sha1.New()
	io.WriteString(hasher, baseUrl)
	hash := hex.EncodeToString(hasher.Sum(nil))
	hashedBaseUrl := hash[:20]
	formattedBaseUrl := hashedBaseUrl + "_" + baseUrl
	return formattedBaseUrl
}

func (videoData *VideoData) GetValidLinks(domains []string) ([]DomainPathIdentifier, error) {
	res := []DomainPathIdentifier{}
	pathIdentifier := videoData.GetUrlPathUniqueIdentifier()
	type GetUrlResponse struct {
		url        DomainPathIdentifier
		successful bool
	}
	ch := make(chan GetUrlResponse)
	httpGetReturns200 := func(url string) bool {
		resp, err := http.Get(url)
		if err != nil {
			return false
		}
		return resp.StatusCode == http.StatusOK
	}
	checkDpiAsync := func(dpi DomainPathIdentifier, ch chan<- GetUrlResponse) {
		if httpGetReturns200(dpi.GetIndexDvrUrl()) {
			ch <- GetUrlResponse{url: dpi, successful: true}
		} else {
			ch <- GetUrlResponse{successful: false}
		}
	}
	for _, domain := range domains {
		go checkDpiAsync(DomainPathIdentifier{domain, pathIdentifier}, ch)
	}
	for range domains {
		domainResponse := <-ch
		if domainResponse.successful {
			res = append(res, domainResponse.url)
		}
	}
	return res, nil
}

type DomainPathIdentifier struct {
	domain        string // e.g. https://d1m7jfoe9zdc1j.cloudfront.net/
	pathIdentifer string // e.g. d138758a032739b16ab9_goonergooch_46448856909_1653186703
}

func (d *DomainPathIdentifier) GetIndexDvrUrl() string {
	return d.domain + d.pathIdentifer + "/chunked/index-dvr.m3u8"
}

func (d *DomainPathIdentifier) GetSegmentChunkedUrl(segment *m3u8.MediaSegment) string {
	return d.domain + d.pathIdentifer + "/chunked/" + segment.URI
}

func (d *DomainPathIdentifier) MakePathsExplicit(playlist *m3u8.MediaPlaylist) *m3u8.MediaPlaylist {
	for _, segment := range playlist.Segments {
		if segment == nil {
			continue
		}
		segment.URI = d.GetSegmentChunkedUrl(segment)
	}
	return playlist
}

func DecodeMediaPlaylist(reader io.Reader, strict bool) (*m3u8.MediaPlaylist, error) {
	p, listType, err := m3u8.DecodeFrom(reader, strict)
	if err != nil {
		return nil, err
	}
	if listType != m3u8.MEDIA {
		return nil, errors.New("m3u8 is not media type")
	}
	return p.(*m3u8.MediaPlaylist), nil
}

func MuteMediaSegments(playlist *m3u8.MediaPlaylist) []*m3u8.MediaSegment {
	nonnilSegments := lo.Filter(playlist.Segments, func(segment *m3u8.MediaSegment, index int) bool {
		return segment != nil
	})
	lo.ForEach(nonnilSegments, func(segment *m3u8.MediaSegment, index int) {
		getNewURI := func(val string) string {
			if strings.Contains(val, "unmuted") {
				start := strings.Index(val, "-")
				front := val[0:start]
				return front + "-muted.ts"
			}
			return val
		}
		segment.URI = getNewURI(segment.URI)
	})
	return nonnilSegments
}

func FetchMediaPlaylist(url string) (*m3u8.MediaPlaylist, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	mediapl, err := DecodeMediaPlaylist(res.Body, true)
	if err != nil {
		return nil, err
	}
	return mediapl, err
}
