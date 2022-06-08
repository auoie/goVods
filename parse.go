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
	"sync"
	"time"

	"github.com/grafov/m3u8"
	"github.com/samber/lo"
)

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

// https://stackoverflow.com/questions/58427586/how-to-return-first-http-response-to-answer
func GetFirstValidDpi(dpis []DomainPathIdentifier) (*DomainPathIdentifier, error) {
	wg := sync.WaitGroup{}
	ch := make(chan *DomainPathIdentifier, len(dpis))
	for _, dpi := range dpis {
		wg.Add(1)
		go func(dpi DomainPathIdentifier) {
			defer wg.Done()
			resp, err := http.Get(dpi.GetIndexDvrUrl())
			if err != nil {
				return
			}
			if resp.StatusCode == http.StatusOK {
				ch <- &dpi
			}
		}(dpi)
	}
	go func() {
		wg.Wait()
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
