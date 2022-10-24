package vods

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	firstnonerr "github.com/auoie/firstNonErr"
	"github.com/grafov/m3u8"
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

type urlIndexResponse struct {
	index int
	valid bool
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

func (domainWithPaths *DomainWithPaths) GetFirstValidDWP(ctx context.Context) (*ValidDwpResponse, error) {
	return firstnonerr.GetFirstNonError(
		ctx,
		domainWithPaths.ToListOfDomainWithPath(),
		0,
		func(ctx context.Context, item *DomainWithPath) (*ValidDwpResponse, error) {
			body, err := item.GetM3U8Body(ctx)
			return &ValidDwpResponse{Dwp: item, Body: body}, err
		})
}

func GetFirstValidDwp(ctx context.Context, domainWithPathsList []*DomainWithPaths) (*ValidDwpResponse, error) {
	return firstnonerr.GetFirstNonError(
		ctx,
		domainWithPathsList,
		0,
		func(ctx context.Context, item *DomainWithPaths) (*ValidDwpResponse, error) {
			return item.GetFirstValidDWP(ctx)
		})
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
		segment.URI = d.GetSegmentChunkedUrl(segment)
	}
	return playlist
}

func (d *DomainWithPath) GetM3U8Body(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.GetIndexDvrUrl(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprint("status code is ", resp.StatusCode))
	}
	return io.ReadAll(resp.Body)
}

func DecodeMediaPlaylistFilterNilSegments(data []byte, strict bool) (*m3u8.MediaPlaylist, error) {
	p, listType, err := m3u8.Decode(*bytes.NewBuffer(data), strict)
	if err != nil {
		return nil, err
	}
	if listType != m3u8.MEDIA {
		return nil, errors.New("m3u8 is not media type")
	}
	mediapl := p.(*m3u8.MediaPlaylist)
	segments := []*m3u8.MediaSegment{}
	for _, segment := range mediapl.Segments {
		if segment != nil {
			segments = append(segments, segment)
		}
	}
	mediapl.Segments = segments
	return mediapl, nil
}

func getMutedURI(segmentUri string) string {
	if strings.Contains(segmentUri, "unmuted") {
		start := strings.Index(segmentUri, "-")
		front := segmentUri[0:start]
		return front + "-muted.ts"
	}
	return segmentUri
}

func MuteMediaSegments(playlist *m3u8.MediaPlaylist) []*m3u8.MediaSegment {
	nonnilSegments := []*m3u8.MediaSegment{}
	for _, segment := range playlist.Segments {
		segment.URI = getMutedURI(segment.URI)
		nonnilSegments = append(nonnilSegments, segment)
	}
	return nonnilSegments
}

func GetMediaPlaylistDuration(mediapl *m3u8.MediaPlaylist) time.Duration {
	duration := 0.0
	for _, segment := range mediapl.Segments {
		duration += segment.Duration
	}
	return time.Duration(duration * float64(time.Second))
}

func GetValidSegments(mediapl *m3u8.MediaPlaylist) []*m3u8.MediaSegment {
	urls := []string{}
	for _, segment := range mediapl.Segments {
		urls = append(urls, segment.URI)
	}
	sortedValidIndices := getSortedIndicesOfValidUrls(urls)
	segments := []*m3u8.MediaSegment{}
	for _, validIndex := range sortedValidIndices {
		segments = append(segments, mediapl.Segments[validIndex])
	}
	return segments
}

func GetMediaPlaylistWithValidSegments(rawPlaylist *m3u8.MediaPlaylist) (*m3u8.MediaPlaylist, error) {
	validSegments := GetValidSegments(rawPlaylist)
	numValidSegments := uint(len(validSegments))
	mediapl, err := m3u8.NewMediaPlaylist(rawPlaylist.WinSize(), numValidSegments)
	if err != nil {
		return nil, err
	}
	for _, validSegment := range validSegments {
		mediapl.AppendSegment(validSegment)
	}
	mediapl.TargetDuration = rawPlaylist.TargetDuration
	mediapl.MediaType = rawPlaylist.MediaType
	mediapl.Closed = rawPlaylist.Closed
	return mediapl, err
}

func getSortedIndicesOfValidUrls(urls []string) []int {
	validIndices := []int{}
	validIndicesCh := make(chan urlIndexResponse)
	for index, url := range urls {
		go func(index int, url string) {
			if urlIsValid(url) {
				validIndicesCh <- urlIndexResponse{index: index, valid: true}
			} else {
				validIndicesCh <- urlIndexResponse{index: index, valid: false}
			}
		}(index, url)
	}
	for range urls {
		response := <-validIndicesCh
		if response.valid {
			validIndices = append(validIndices, response.index)
		}
	}
	sort.Slice(validIndices, func(i, j int) bool {
		return validIndices[i] < validIndices[j]
	})
	return validIndices
}

func urlIsValid(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}
