package vods

import (
	"bytes"
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

var DOMAINS = []string{
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

type VideoData struct {
	StreamerName string
	VideoId      string
	Time         time.Time
}

type videoPath struct {
	urlPath   string // e.g. {hash}_{streamername}_{videoid}_{unixtime}
	videoData *VideoData
}
type DomainWithPath struct {
	domain string // e.g. https://d1m7jfoe9zdc1j.cloudfront.net/
	path   *videoPath
}

type DomainWithPaths struct {
	domain string // e.g. https://d1m7jfoe9zdc1j.cloudfront.net/
	paths  []*videoPath
}

type ValidDwpResponse struct {
	Dwp  *DomainWithPath
	Body []byte
}

type dwpResponse struct {
	validResponse *ValidDwpResponse
	valid         bool
}

func (videoData *VideoData) String() string {
	values := []string{videoData.StreamerName, videoData.Time.Format("2006-01-02_15:04:05"), videoData.VideoId}
	return strings.Join(values, "_")
}

func (videoData *VideoData) GetVideoPath() *videoPath {
	return &videoPath{urlPath: videoData.GetUrlPath(), videoData: videoData}
}

func (videoData *VideoData) GetUrlPath() string {
	unixTime := videoData.Time.Unix()
	baseUrl := videoData.StreamerName + "_" + videoData.VideoId + "_" + fmt.Sprint(unixTime)
	hasher := sha1.New()
	io.WriteString(hasher, baseUrl)
	hash := hex.EncodeToString(hasher.Sum(nil))
	hashedBaseUrl := hash[:20]
	formattedBaseUrl := hashedBaseUrl + "_" + baseUrl
	return formattedBaseUrl
}

func (videoData *VideoData) WithOffset(seconds int) *VideoData {
	return &VideoData{
		StreamerName: videoData.StreamerName,
		VideoId:      videoData.VideoId,
		Time:         videoData.Time.Add(time.Second * time.Duration(seconds)),
	}
}

func (videoData *VideoData) GetDomainWithPathsList(domains []string, seconds int) []*DomainWithPaths {
	videoPaths := []*videoPath{}
	for i := 0; i < seconds; i++ {
		videoPaths = append(videoPaths, videoData.WithOffset(i).GetVideoPath())
	}
	domainWithPathsList := []*DomainWithPaths{}
	for _, domain := range domains {
		domainWithPathsList = append(domainWithPathsList, &DomainWithPaths{domain: domain, paths: videoPaths})
	}
	return domainWithPathsList
}

func (domainWithPaths *DomainWithPaths) ToListOfDomainWithPath() []*DomainWithPath {
	result := []*DomainWithPath{}
	domain := domainWithPaths.domain
	for _, path := range domainWithPaths.paths {
		result = append(result, &DomainWithPath{domain: domain, path: path})
	}
	return result
}

func (domainWithPaths *DomainWithPaths) GetFirstValidDWP() (*ValidDwpResponse, error) {
	domainWithPathList := domainWithPaths.ToListOfDomainWithPath()
	if len(domainWithPathList) < 1 {
		return nil, errors.New("no urls")
	}
	// establish TCP connection for reuse
	// https://groups.google.com/g/golang-nuts/c/5T5aiDRl_cw/m/zYPGtCOYBwAJ
	firstDomainWithPath := domainWithPathList[0]
	body, err := firstDomainWithPath.GetM3U8Body()
	if err == nil {
		return &ValidDwpResponse{Dwp: firstDomainWithPath, Body: body}, nil
	}
	// reuse with other requests
	restDomainWithPathList := domainWithPathList[1:]
	firstValidResponseCh := make(chan *ValidDwpResponse)
	responsesCh := make(chan *dwpResponse)
	for _, dwp := range restDomainWithPathList {
		go func(dwp *DomainWithPath) {
			body, err := dwp.GetM3U8Body()
			if err == nil {
				responsesCh <- &dwpResponse{validResponse: &ValidDwpResponse{Dwp: dwp, Body: body}, valid: true}
			} else {
				responsesCh <- &dwpResponse{validResponse: nil, valid: false}
			}
		}(dwp)
	}
	go func() {
		for range restDomainWithPathList {
			dwpResponse := <-responsesCh
			if dwpResponse.valid {
				firstValidResponseCh <- dwpResponse.validResponse
				return
			}
		}
		close(firstValidResponseCh)
	}()
	result, ok := <-firstValidResponseCh
	if !ok {
		return nil, errors.New("no valid links were found")
	}
	return result, nil
}

func GetFirstValidDwp(domainWithPathsList []*DomainWithPaths) (*ValidDwpResponse, error) {
	firstValidResponseCh := make(chan *ValidDwpResponse)
	responsesCh := make(chan *dwpResponse)
	for _, domainWithPaths := range domainWithPathsList {
		go func(domainWithPaths *DomainWithPaths) {
			validDwpResponse, err := domainWithPaths.GetFirstValidDWP()
			if err == nil {
				responsesCh <- &dwpResponse{validResponse: validDwpResponse, valid: true}
			} else {
				responsesCh <- &dwpResponse{validResponse: nil, valid: false}
			}
		}(domainWithPaths)
	}
	go func() {
		for range domainWithPathsList {
			response := <-responsesCh
			if response.valid {
				firstValidResponseCh <- response.validResponse
				return
			}
		}
		close(firstValidResponseCh)
	}()
	result, ok := <-firstValidResponseCh
	if !ok {
		return nil, errors.New("no valid links were found")
	}
	return result, nil
}

func (d *DomainWithPath) GetVideoData() *VideoData {
	return d.path.videoData
}

func (d *DomainWithPath) GetIndexDvrUrl() string {
	return d.domain + d.path.urlPath + "/chunked/index-dvr.m3u8"
}

func (d *DomainWithPath) GetSegmentChunkedUrl(segment *m3u8.MediaSegment) string {
	return d.domain + d.path.urlPath + "/chunked/" + segment.URI
}

func (d *DomainWithPath) MakePathsExplicit(playlist *m3u8.MediaPlaylist) *m3u8.MediaPlaylist {
	for _, segment := range playlist.Segments {
		if segment == nil {
			continue
		}
		segment.URI = d.GetSegmentChunkedUrl(segment)
	}
	return playlist
}

func (d *DomainWithPath) GetM3U8Body() ([]byte, error) {
	resp, err := http.Get(d.GetIndexDvrUrl())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprint("status code is ", resp.StatusCode))
	}
	return io.ReadAll(resp.Body)
}

func DecodeMediaPlaylist(data []byte, strict bool) (*m3u8.MediaPlaylist, error) {
	p, listType, err := m3u8.Decode(*bytes.NewBuffer(data), strict)
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

func GetMediaPlaylistDuration(mediapl *m3u8.MediaPlaylist) time.Duration {
	duration := 0.0
	for _, segment := range mediapl.Segments {
		if segment != nil {
			duration += segment.Duration
		}
	}
	return time.Duration(duration * float64(time.Second))
}
