package vods

import "time"

type StreamsChartsData struct {
	StreamerName string
	VideoId      string
	UtcTime      string
}

func (data *StreamsChartsData) GetVideoData() (VideoData, error) {
	time, err := time.Parse("02-01-2006 15:04", data.UtcTime)
	if err != nil {
		return VideoData{}, err
	}
	return VideoData{
		StreamerName: data.StreamerName,
		VideoId:      data.VideoId,
		Time:         time,
	}, nil
}
