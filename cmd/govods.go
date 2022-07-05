package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/auoie/goVods"
	"github.com/urfave/cli/v2"
)

func mainHelper(dpis []goVods.DomainPathIdentifier, ctx *cli.Context) error {
	write := ctx.Bool("write")
	dpi, err := goVods.GetFirstValidDpi(dpis)
	if err != nil {
		return err
	}
	mediapl, err := goVods.FetchMediaPlaylist(dpi.GetIndexDvrUrl())
	if err != nil {
		return err
	}
	goVods.MuteMediaSegments(mediapl)
	dpi.MakePathsExplicit(mediapl)
	if write {
		videoData, err := dpi.ToVideoData()
		if err != nil {
			return err
		}
		directoryPath := filepath.Join("Downloads", videoData.StreamerName)
		if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
			return err
		}
		roundedDuration := goVods.GetMediaPlaylistDuration(mediapl).Truncate(time.Second)
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
	} else {
		fmt.Println(mediapl.String())
	}
	return nil
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
					&cli.BoolFlag{
						Name:  "write",
						Usage: "Rather than printing the file, write the .m3u8 file to the folder ./Downloads/.",
					},
				},
				Action: func(ctx *cli.Context) error {
					streamer := ctx.String("streamer")
					videoid := ctx.String("videoid")
					time := ctx.String("time")
					twitchData := goVods.TwitchTrackerData{StreamerName: streamer, VideoId: videoid, UtcTime: time}
					videoData, err := twitchData.GetVideoData()
					if err != nil {
						return err
					}
					dpis := videoData.GetDpis(goVods.DOMAINS)
					return mainHelper(dpis, ctx)
				},
			},
			{
				Name:  "tt-url-get-m3u8",
				Usage: "Using a twitchtracker.com url, get an .m3u8 file which can be viewed in a media player.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "url",
						Usage:    "Twitch tracker URL for the Twitch stream",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "write",
						Usage: "Rather than printing the file, write the .m3u8 file to the folder ./Downloads/.",
					},
					&cli.StringFlag{
						Name:  "streamer",
						Usage: "If a streamer changes their username, you need to provide their old username for their vods created before the name change.",
					},
				},
				Action: func(ctx *cli.Context) error {
					twitchTrackerUrl := ctx.String("url")
					twitchData, err := goVods.GetTwitchTrackerData(twitchTrackerUrl)
					if err != nil {
						return err
					}
					videoData, err := twitchData.GetVideoData()
					if err != nil {
						return err
					}
					streamer := ctx.String("streamer")
					if streamer != "" {
						videoData.StreamerName = streamer
					}
					dpis := videoData.GetDpis(goVods.DOMAINS)
					return mainHelper(dpis, ctx)
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
					&cli.BoolFlag{
						Name:  "write",
						Usage: "Rather than printing the file, write the .m3u8 file to the folder ./Downloads/.",
					},
				},
				Action: func(ctx *cli.Context) error {
					streamer := ctx.String("streamer")
					videoid := ctx.String("videoid")
					time := ctx.String("time")
					scData := goVods.StreamsChartsData{StreamerName: streamer, VideoId: videoid, UtcTime: time}
					videoData, err := scData.GetVideoData()
					if err != nil {
						return err
					}
					dpis := []goVods.DomainPathIdentifier{}
					for i := 0; i < 60; i++ {
						dpis = append(dpis, videoData.WithOffset(i).GetDpis(goVods.DOMAINS)...)
					}
					return mainHelper(dpis, ctx)
				},
			},
			{
				Name:  "sc-url-get-m3u8",
				Usage: "Using a streamscharts.com url, get an .m3u8 file which can be viewed in a media player.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "url",
						Usage:    "Streams Charts URL for the Twitch stream",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "write",
						Usage: "Rather than printing the file, write the .m3u8 file to the folder ./Downloads/.",
					},
					&cli.StringFlag{
						Name:  "streamer",
						Usage: "If a streamer changes their username, you need to provide their old username for their vods created before the name change.",
					},
				},
				Action: func(ctx *cli.Context) error {
					streamsChartsUrl := ctx.String("url")
					scData, err := goVods.GetStreamsChartsData(streamsChartsUrl)
					if err != nil {
						return err
					}
					videoData, err := scData.GetVideoData()
					if err != nil {
						return err
					}
					streamer := ctx.String("streamer")
					if streamer != "" {
						videoData.StreamerName = streamer
					}
					dpis := []goVods.DomainPathIdentifier{}
					for i := 0; i < 60; i++ {
						dpis = append(dpis, videoData.WithOffset(i).GetDpis(goVods.DOMAINS)...)
					}
					return mainHelper(dpis, ctx)
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
