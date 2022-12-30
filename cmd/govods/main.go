package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/auoie/goVods/vods"
	"github.com/grafov/m3u8"
	"github.com/urfave/cli/v2"
)

func writeMediaPlaylist(mediapl *m3u8.MediaPlaylist, dpi *vods.ValidDwpResponse) error {
	videoData := dpi.Dwp.GetVideoData()
	directoryPath := filepath.Join("Downloads", videoData.StreamerName)
	if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
		return err
	}
	roundedDuration := vods.GetMediaPlaylistDuration(mediapl).Truncate(time.Second)
	filePath := filepath.Join(directoryPath, fmt.Sprint(videoData, "_", roundedDuration, ".m3u8"))
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, mediapl.Encode())
	if err != nil {
		return err
	}
	return nil
}

func makeRobustClient() *http.Client {
	timeout := 10 * time.Second
	dialer := &net.Dialer{
		Timeout: timeout,
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: &http.Transport{DialContext: dialer.DialContext},
	}
}

func getValidDwp(ctx context.Context, domains []string, seconds int, videoData *vods.VideoData, client *http.Client) (*vods.ValidDwpResponse, error) {
	domainWithPathsList := videoData.GetDomainWithPathsList(vods.DOMAINS, seconds, true)
	dwpAndBody, err := vods.GetFirstValidDwp(ctx, domainWithPathsList, client)
	if err == nil {
		return dwpAndBody, nil
	}
	// very rarely, a stream will use the seconds of the time rather than the unix time in the m3u8 file name
	domainWithPathsList = videoData.GetDomainWithPathsList(vods.DOMAINS, seconds, false)
	dwpAndBody, err = vods.GetFirstValidDwp(ctx, domainWithPathsList, client)
	if err == nil {
		return dwpAndBody, nil
	}
	return nil, err
}

func mainHelper(seconds int, videoData *vods.VideoData, ctx *cli.Context) error {
	videoData = videoData.WithOffset(-1) // some m3u8 file names use a time that is 1 second minus the provided time
	client := makeRobustClient()
	dwpAndBody, err := getValidDwp(ctx.Context, vods.DOMAINS, seconds+1, videoData, client)
	if err != nil {
		return err
	}
	fmt.Println(fmt.Sprint("Found valid url ", dwpAndBody.Dwp.GetIndexDvrUrl()))
	mediapl, err := vods.DecodeMediaPlaylistFilterNilSegments(dwpAndBody.Body, true)
	if err != nil {
		return err
	}
	vods.MuteMediaSegments(mediapl)
	dwpAndBody.Dwp.MakePathsExplicit(mediapl)
	checkInvalidConcurrent := ctx.Int("filter-invalid")
	if checkInvalidConcurrent > 0 {
		numTotalSegments := len(mediapl.Segments)
		mediapl, err = vods.GetMediaPlaylistWithValidSegments(mediapl, checkInvalidConcurrent, client)
		if err != nil {
			return err
		}
		numValidSegments := len(mediapl.Segments)
		fmt.Println(fmt.Sprint(numValidSegments, " valid segments out of ", numTotalSegments))
		if numValidSegments == 0 {
			return errors.New("0 valid segments found")
		}
	}
	return writeMediaPlaylist(mediapl, dwpAndBody)
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "tt-manual-get-m3u8",
				Usage: "Using twitchtracker.com data, get an .m3u8 file which can be viewed in a media player.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "streamer",
						Usage:    "twitch streamer name",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "videoid",
						Usage:    "twitch tracker video id",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "time",
						Usage:    "stream UTC start time in the format '2006-01-02 15:04:05' (year-month-day hour:minute:second)",
						Required: true,
					},
					&cli.IntFlag{
						Name:  "filter-invalid",
						Usage: "Filter out all of the invalid segments in the m3u8 file with concurrency level",
					},
				},
				Action: func(ctx *cli.Context) error {
					streamer := ctx.String("streamer")
					videoid := ctx.String("videoid")
					time := ctx.String("time")
					twitchData := vods.TwitchTrackerData{StreamerName: streamer, VideoId: videoid, UtcTime: time}
					videoData, err := twitchData.GetVideoData()
					if err != nil {
						return err
					}
					return mainHelper(1, &videoData, ctx)
				},
			},
			{
				Name:  "sc-manual-get-m3u8",
				Usage: "Using streamscharts.com data, get an .m3u8 file which can be viewed in a media player.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "streamer",
						Usage:    "twitch streamer name",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "videoid",
						Usage:    "twitch tracker video id",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "time",
						Usage:    "stream UTC start time in the format '02-01-2006 15:04' (day-month-year hour:minute)",
						Required: true,
					},
					&cli.IntFlag{
						Name:  "filter-invalid",
						Usage: "Filter out all of the invalid segments in the m3u8 file with concurrency level",
					},
				},
				Action: func(ctx *cli.Context) error {
					streamer := ctx.String("streamer")
					videoid := ctx.String("videoid")
					time := ctx.String("time")
					scData := vods.StreamsChartsData{StreamerName: streamer, VideoId: videoid, UtcTime: time}
					videoData, err := scData.GetVideoData()
					if err != nil {
						return err
					}
					return mainHelper(60, &videoData, ctx)
				},
			},
			{
				Name:  "sg-manual-get-m3u8",
				Usage: "Using sullygnome.com data, get an .m3u8 file which can be viewed in a media player.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "streamer",
						Usage:    "twitch streamer name",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "videoid",
						Usage:    "twitch tracker video id",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "time",
						Usage:    "stream UTC start time in the format '2006-01-02T15:04:05Z' (year-month-dayThour:minute:secondZ)",
						Required: true,
					},
					&cli.IntFlag{
						Name:  "filter-invalid",
						Usage: "Filter out all of the invalid segments in the m3u8 file with concurrency level",
					},
				},
				Action: func(ctx *cli.Context) error {
					streamer := ctx.String("streamer")
					videoid := ctx.String("videoid")
					time := ctx.String("time")
					sullygnomeData := vods.SullyGnomeData{StreamerName: streamer, VideoId: videoid, UtcTime: time}
					videoData, err := sullygnomeData.GetVideoData()
					if err != nil {
						return err
					}
					return mainHelper(1, &videoData, ctx)
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
