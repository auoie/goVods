package govods

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

var (
	DOMAINS = []string{
		"https://vod-secure.twitch.tv/",
		"https://vod-metro.twitch.tv/",
		"https://vod-pop-secure.twitch.tv/",
		"https://d2e2de1etea730.cloudfront.net/",
		"https://dqrpb9wgowsf5.cloudfront.net/",
		"https://ds0h3roq6wcgc.cloudfront.net/",
		"https://d2nvs31859zcd8.cloudfront.net/",
		"https://d2aba1wr3818hz.cloudfront.net/",
		"https://d3c27h4odz752x.cloudfront.net/",
		"https://dgeft87wbj63p.cloudfront.net/",
		"https://d1m7jfoe9zdc1j.cloudfront.net/",
		"https://d3vd9lfkzbru3h.cloudfront.net/",
		"https://d2vjef5jvl6bfs.cloudfront.net/",
		"https://d1ymi26ma8va5x.cloudfront.net/",
		"https://d1mhjrowxxagfy.cloudfront.net/",
		"https://ddacn6pr5v0tl.cloudfront.net/",
		"https://d3aqoihi2n8ty8.cloudfront.net/",
	}
)

type TwitchTrackerData struct {
	StreamerName string
	VideoId      int
	UtcTime      string
}

func (ttData *TwitchTrackerData) getVideoData() (VideoData, error) {
	time, err := time.Parse("2006-01-2 15:04:05", ttData.UtcTime)
	if err != nil {
		return VideoData{}, err
	}
	return VideoData{
		streamerName: ttData.StreamerName,
		videoId:      ttData.VideoId,
		time:         time,
	}, nil
}

type VideoData struct {
	streamerName string
	videoId      int
	time         time.Time
}

func (videoData *VideoData) getUrlPathUniqueIdentifier() string {
	unixTime := videoData.time.Unix()
	baseUrl := videoData.streamerName + "_" + fmt.Sprint(videoData.videoId) + "_" + fmt.Sprint(unixTime)
	hasher := sha1.New()
	io.WriteString(hasher, baseUrl)
	hash := hex.EncodeToString(hasher.Sum(nil))
	hashedBaseUrl := hash[:20]
	formattedBaseUrl := hashedBaseUrl + "_" + baseUrl
	return formattedBaseUrl
}

type DomainPathIdentifier struct {
	domain        string // e.g. https://d1m7jfoe9zdc1j.cloudfront.net/
	pathIdentifer string // e.g. d138758a032739b16ab9_goonergooch_46448856909_1653186703
}

func (d *DomainPathIdentifier) GetIndexUrl() string {
	return d.domain + d.pathIdentifer + "/chunked/index-dvr.m3u8"
}

func (d *DomainPathIdentifier) getChunkUrl(uri string) string {
	return d.domain + d.pathIdentifer + "/chunked/" + uri

}

type GetUrlResponse struct {
	url        DomainPathIdentifier
	successful bool
}

func httpGetReturns200(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	return resp.StatusCode == http.StatusOK
}

func checkUrlAsync(url DomainPathIdentifier, ch chan<- GetUrlResponse) {
	if httpGetReturns200(url.GetIndexUrl()) {
		ch <- GetUrlResponse{url: url, successful: true}
	} else {
		ch <- GetUrlResponse{successful: false}
	}
}

func GetValidLinks(ttData TwitchTrackerData) ([]DomainPathIdentifier, error) {
	videoData, err := ttData.getVideoData()
	if err != nil {
		return nil, err
	}
	res := []DomainPathIdentifier{}
	pathIdentifier := videoData.getUrlPathUniqueIdentifier()
	ch := make(chan GetUrlResponse)
	for _, domain := range DOMAINS {
		go checkUrlAsync(DomainPathIdentifier{domain, pathIdentifier}, ch)
	}
	for range DOMAINS {
		domainResponse := <-ch
		if domainResponse.successful {
			res = append(res, domainResponse.url)
		}
	}
	return res, nil
}

func decodeMediaPlaylist(reader io.Reader, strict bool) (*m3u8.MediaPlaylist, error) {
	p, listType, err := m3u8.DecodeFrom(reader, strict)
	if err != nil {
		return nil, err
	}
	if listType != m3u8.MEDIA {
		return nil, errors.New("m3u8 is not media type")
	}
	return p.(*m3u8.MediaPlaylist), nil
}

func getChunkUris(dpi DomainPathIdentifier) ([]string, error) {
	res, err := http.Get(dpi.GetIndexUrl())
	if err != nil {
		return nil, err
	}
	mediapl, err := decodeMediaPlaylist(res.Body, true)
	if err != nil {
		return nil, err
	}
	unfilteredUris := lo.Map(mediapl.Segments, func(segment *m3u8.MediaSegment, index int) string {
		if segment == nil {
			return ""
		}
		return segment.URI
	})
	rawUris := lo.Filter(unfilteredUris, func(val string, index int) bool {
		return val != ""
	})
	processedUris := lo.Map(rawUris, func(val string, index int) string {
		if strings.Contains(val, "unmuted") {
			start := strings.Index(val, "-")
			front := val[0:start]
			return front + "-muted.ts"
		}
		return val
	})
	return processedUris, nil
}
