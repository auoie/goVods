package goVods

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/samber/lo"
)

type StreamsChartsData struct {
	StreamerName string
	VideoId      string
	UtcTime      string
}

func (data *StreamsChartsData) GetVideoData() (VideoData, error) {
	time, err := time.Parse("02-01-2006 15:04", data.UtcTime)
	if err != nil {
		return VideoData{}, nil
	}
	return VideoData{
		StreamerName: data.StreamerName,
		VideoId:      data.VideoId,
		Time:         time,
	}, nil
}

func streamsChartsUrlGetNameAndId(streamsChartsUrl string) (streamerNameAndVideoId, error) {
	u, err := url.Parse(streamsChartsUrl)
	if err != nil {
		return streamerNameAndVideoId{}, err
	}
	parts := lo.Filter(strings.Split(u.Path, "/"), func(val string, index int) bool { return val != "" })
	if len(parts) < 4 {
		return streamerNameAndVideoId{}, errors.New(fmt.Sprint("the path segments were ", parts))
	}
	return streamerNameAndVideoId{StreamerName: parts[1], VideoId: parts[3]}, nil
}

func streamsChartsUrlGetUtcTime(streamsChartsUrl string, streamerName string) (string, error) {
	req, err := http.NewRequest("GET", streamsChartsUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authority", "streamscharts.com")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", fmt.Sprint("https://streamscharts.com/channels/", streamerName, "/streams"))
	req.Header.Set("Sec-Ch-Ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"102\", \"Google Chrome\";v=\"102\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "\"Linux\"")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprint("status not ok: ", resp.Status))
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", nil
	}
	s := doc.Find("time").First()
	if s == nil {
		return "", errors.New("could not find <time></time> element")
	}
	result, exists := s.Attr("datetime")
	if !exists {
		return "", errors.New("datetime attribute does not exist in the <time></time> element")
	}
	return result, nil
}

func GetStreamsChartsData(streamsChartsUrl string) (StreamsChartsData, error) {
	nameAndId, err := streamsChartsUrlGetNameAndId(streamsChartsUrl)
	if err != nil {
		return StreamsChartsData{}, err
	}
	utctime, err := streamsChartsUrlGetUtcTime(streamsChartsUrl, nameAndId.StreamerName)
	if err != nil {
		return StreamsChartsData{}, err
	}
	return StreamsChartsData{StreamerName: nameAndId.StreamerName, VideoId: nameAndId.VideoId, UtcTime: utctime}, nil
}
