package goVods

import "time"

type SullyGnomeData struct {
	StreamerName string
	VideoId      string
	UtcTime      string
}

func (data *SullyGnomeData) GetVideoData() (VideoData, error) {
	time, err := time.Parse("2006-01-02T15:04:05Z", data.UtcTime)
	if err != nil {
		return VideoData{}, err
	}
	return VideoData{
		StreamerName: data.StreamerName,
		VideoId:      data.VideoId,
		Time:         time,
	}, nil
}
