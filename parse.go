package goVods

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
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

func (videoData *VideoData) String() string {
	values := []string{videoData.StreamerName, videoData.Time.Format("2006-01-02_15:04:05"), videoData.VideoId}
	return strings.Join(values, "_")
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

func (videoData *VideoData) WithOffset(seconds int) *VideoData {
	return &VideoData{
		StreamerName: videoData.StreamerName,
		VideoId:      videoData.VideoId,
		Time:         videoData.Time.Add(time.Second * time.Duration(seconds)),
	}
}

func (videoData *VideoData) GetDpi(domain string) DomainPathIdentifier {
	pathIdentifier := videoData.GetUrlPathUniqueIdentifier()
	return DomainPathIdentifier{domain: domain, pathIdentifer: pathIdentifier}
}

func (videoData *VideoData) GetDpis(domains []string) []DomainPathIdentifier {
	res := []DomainPathIdentifier{}
	for _, domain := range domains {
		res = append(res, videoData.GetDpi(domain))
	}
	return res
}

type DpiResponse struct {
	dpi   *DomainPathIdentifier
	valid bool
}

func GetFirstValidDpi(dpis []DomainPathIdentifier) (*DomainPathIdentifier, error) {
	ch := make(chan *DomainPathIdentifier)
	responsesCh := make(chan *DpiResponse)
	for _, dpi := range dpis {
		go func(dpi DomainPathIdentifier) {
			resp, err := http.Get(dpi.GetIndexDvrUrl())
			if err == nil && resp.StatusCode == http.StatusOK {
				responsesCh <- &DpiResponse{dpi: &dpi, valid: true}
			} else {
				responsesCh <- &DpiResponse{dpi: &dpi, valid: false}
			}
		}(dpi)
	}
	go func() {
		for range dpis {
			dpiResponse := <-responsesCh
			if dpiResponse.valid {
				ch <- dpiResponse.dpi
				return
			}
		}
		close(ch)
	}()
	result, ok := <-ch
	if !ok {
		return nil, errors.New("no valid links were found")
	}
	return result, nil
}

type DomainPathIdentifier struct {
	domain        string // e.g. https://d1m7jfoe9zdc1j.cloudfront.net/
	pathIdentifer string // e.g. {hash}_{streamername}_{videoid}_{unixtime}
}

func (d *DomainPathIdentifier) ToVideoData() (*VideoData, error) {
	allUnderscoreIndices := []int{}
	for i := 0; i < len(d.pathIdentifer); i++ {
		char := d.pathIdentifer[i]
		if char == '_' {
			allUnderscoreIndices = append(allUnderscoreIndices, i)
		}
	}
	numUnderscores := len(allUnderscoreIndices)
	if numUnderscores < 3 {
		return nil, errors.New("the pathIdentifer doesn't have enough enough underscores")
	}
	underscoreIndices := [3]int{allUnderscoreIndices[0], allUnderscoreIndices[numUnderscores-2], allUnderscoreIndices[numUnderscores-1]}
	streamerName := d.pathIdentifer[underscoreIndices[0]+1 : underscoreIndices[1]]
	videoid := d.pathIdentifer[underscoreIndices[1]+1 : underscoreIndices[2]]
	unixtimeString := d.pathIdentifer[underscoreIndices[2]+1:]
	unixtimeInt, err := strconv.ParseInt(unixtimeString, 10, 64)
	if err != nil {
		return nil, err
	}
	videoTime := time.Unix(unixtimeInt, 0)
	return &VideoData{StreamerName: streamerName, VideoId: videoid, Time: videoTime}, nil
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
	return DecodeMediaPlaylist(res.Body, true)
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
