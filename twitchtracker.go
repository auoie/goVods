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

func (ttData *TwitchTrackerData) GetVideoData() (VideoData, error) {
	time, err := time.Parse("2006-01-02 15:04:05", ttData.UtcTime)
	if err != nil {
		return VideoData{}, err
	}
	return VideoData{
		StreamerName: ttData.StreamerName,
		VideoId:      ttData.VideoId,
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

func twitchTrackerUrlGetUtcTime(twitchTrackerUrl string) (string, error) {
	req, err := http.NewRequest("GET", twitchTrackerUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", "curl/7.83.1")
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
	utctime, err := twitchTrackerUrlGetUtcTime(twitchTrackerUrl)
	if err != nil {
		return TwitchTrackerData{}, err
	}
	return TwitchTrackerData{StreamerName: nameAndId.StreamerName, VideoId: nameAndId.VideoId, UtcTime: utctime}, nil
}
