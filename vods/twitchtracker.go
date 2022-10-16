package vods

import "time"

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
