package vods_test

import (
	"testing"
	"time"

	"github.com/auoie/goVods/vods"
)

func assertEqual[T comparable](t testing.TB, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf(`got %v want %v`, got, want)
	}
}

func TestUrlPathToVideoData(t *testing.T) {
	urlPath := "c5992ececce7bd7d350d_malek_04_47198535725_1664038929"
	result, err := vods.UrlPathToVideoData(urlPath)
	if err != nil {
		t.Fatalf(err.Error())
	}
	want := vods.VideoData{
		StreamerName: "malek_04",
		VideoId:      "47198535725",
		Time:         time.Unix(1664038929, 0),
	}
	assertEqual(t, *result, want)
}

func TestUrlToDomainWithPath(t *testing.T) {
	url := "https://d1m7jfoe9zdc1j.cloudfront.net/c5992ececce7bd7d350d_gmhikaru_47198535725_1664038929/storyboards/1600104857-info.json"
	result, err := vods.UrlToDomainWithPath(url)
	if err != nil {
		t.Fatalf(err.Error())
	}
	assertEqual(t, result.Domain, "https://d1m7jfoe9zdc1j.cloudfront.net/")
	assertEqual(t, *result.Path.VideoData, vods.VideoData{StreamerName: "gmhikaru", VideoId: "47198535725", Time: time.Unix(1664038929, 0)})
	assertEqual(t, result.Path.UrlPath, "c5992ececce7bd7d350d_gmhikaru_47198535725_1664038929")
}
