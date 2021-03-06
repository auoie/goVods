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

type TwitchTrackerData struct {
	StreamerName string
	VideoId      string
	UtcTime      string
}

func (data *TwitchTrackerData) GetVideoData() (VideoData, error) {
	time, err := time.Parse("2006-01-02 15:04:05", data.UtcTime)
	if err != nil {
		return VideoData{}, err
	}
	return VideoData{
		StreamerName: data.StreamerName,
		VideoId:      data.VideoId,
		Time:         time,
	}, nil
}

type streamerNameAndVideoId struct {
	StreamerName string
	VideoId      string
}

func twitchTrackerUrlGetNameAndId(twitchTrackerUrl string) (streamerNameAndVideoId, error) {
	u, err := url.Parse(twitchTrackerUrl)
	if err != nil {
		return streamerNameAndVideoId{}, err
	}
	parts := lo.Filter(strings.Split(u.Path, "/"), func(val string, index int) bool { return val != "" })
	if len(parts) < 3 {
		return streamerNameAndVideoId{}, errors.New(fmt.Sprint("the path segments were ", parts))
	}
	return streamerNameAndVideoId{StreamerName: parts[0], VideoId: parts[2]}, nil
}

func twitchTrackerUrlGetUtcTime(twitchTrackerUrl string, streamerName string) (string, error) {
	req, err := http.NewRequest("GET", twitchTrackerUrl, nil)
	if err != nil {
		return "", nil
	}
	req.Header.Set("Authority", "twitchtracker.com")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", fmt.Sprint("https://twitchtracker.com/", streamerName, "/streams"))
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
		return "", nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprint("status not ok: ", resp.Status))
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", nil
	}
	s := doc.Find(".stream-timestamp-dt").First()
	if s == nil {
		return "", errors.New("could not find element with .stream-timestamp-dt")
	}
	return s.Text(), nil
}

func GetTwitchTrackerData(twitchTrackerUrl string) (TwitchTrackerData, error) {
	nameAndId, err := twitchTrackerUrlGetNameAndId(twitchTrackerUrl)
	if err != nil {
		return TwitchTrackerData{}, err
	}
	utctime, err := twitchTrackerUrlGetUtcTime(twitchTrackerUrl, nameAndId.StreamerName)
	if err != nil {
		return TwitchTrackerData{}, err
	}
	return TwitchTrackerData{StreamerName: nameAndId.StreamerName, VideoId: nameAndId.VideoId, UtcTime: utctime}, nil
}
