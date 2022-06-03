package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	govods "goVods"

	"github.com/urfave/cli/v2"
)

var DOMAINS = []string{
	"https://vod-secure.twitch.tv/",
	"https://vod-metro.twitch.tv/",
	"https://vod-pop-secure.twitch.tv/",
	"https://d2e2de1etea730.cloudfront.net/",
	"https://dqrpb9wgowsf5.cloudfront.net/",
	"https://ds0h3roq6wcgc.cloudfront.net/",
	"https://d2nvs31859zcd8.cloudfront.net/",
	"https://d2aba1wr3818hz.cloudfront.net/",
	"https://d3c27h4odz752x.cloudfront.net/",
	"https://dgeft87wbj63p.cloudfront.net/",
	"https://d1m7jfoe9zdc1j.cloudfront.net/",
	"https://d3vd9lfkzbru3h.cloudfront.net/",
	"https://d2vjef5jvl6bfs.cloudfront.net/",
	"https://d1ymi26ma8va5x.cloudfront.net/",
	"https://d1mhjrowxxagfy.cloudfront.net/",
	"https://ddacn6pr5v0tl.cloudfront.net/",
	"https://d3aqoihi2n8ty8.cloudfront.net/",
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "urls",
				Usage: "Get URLS of media HLS manifests",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "url",
						Usage:    "Twitch tracker URL for the Twitch stream",
						Required: true,
					},
				},
				Action: func(ctx *cli.Context) error {
					twitchTrackerUrl := ctx.String("url")
					twitchData, err := govods.GetTwitchTrackerData(twitchTrackerUrl)
					if err != nil {
						return err
					}
					videoData, err := twitchData.GetVideoData()
					if err != nil {
						return err
					}
					validPathIdentifiers, err := videoData.GetValidLinks(DOMAINS)
					if err != nil {
						return err
					}
					for _, dpi := range validPathIdentifiers {
						fmt.Println(dpi.GetIndexDvrUrl())
					}
					return nil
				},
			},
			{
				Name:  "manual-get-m3u8",
				Usage: "Get m3u8 file which can be viewed in media player",
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
						Usage:    "stream UTC start time in the format '2006-01-02 15:04:05'",
						Required: true,
					},
				},
				Action: func(ctx *cli.Context) error {
					streamer := ctx.String("streamer")
					videoid := ctx.String("videoid")
					time := ctx.String("time")
					twitchData := govods.TwitchTrackerData{StreamerName: streamer, VideoId: videoid, UtcTime: time}
					videoData, err := twitchData.GetVideoData()
					if err != nil {
						return err
					}
					validPathIdentifiers, err := videoData.GetValidLinks(DOMAINS)
					if err != nil {
						return err
					}
					if len(validPathIdentifiers) == 0 {
						return errors.New("no valid urls were found")
					}
					dpi := validPathIdentifiers[0]
					mediapl, err := govods.FetchMediaPlaylist(dpi.GetIndexDvrUrl())
					if err != nil {
						return err
					}
					govods.MuteMediaSegments(mediapl)
					dpi.MakePathsExplicit(mediapl)
					fmt.Println(mediapl.String())
					return nil
				},
			},
			{
				Name:  "get-m3u8",
				Usage: "Get m3u8 file which can be viewed in media player",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "url",
						Usage:    "Twitch tracker URL for the Twitch stream",
						Required: true,
					},
				},
				Action: func(ctx *cli.Context) error {
					twitchTrackerUrl := ctx.String("url")
					twitchData, err := govods.GetTwitchTrackerData(twitchTrackerUrl)
					if err != nil {
						return err
					}
					videoData, err := twitchData.GetVideoData()
					if err != nil {
						return err
					}
					validPathIdentifiers, err := videoData.GetValidLinks(DOMAINS)
					if err != nil {
						return err
					}
					if len(validPathIdentifiers) == 0 {
						return errors.New("no valid urls were found")
					}
					dpi := validPathIdentifiers[0]
					mediapl, err := govods.FetchMediaPlaylist(dpi.GetIndexDvrUrl())
					if err != nil {
						return err
					}
					govods.MuteMediaSegments(mediapl)
					dpi.MakePathsExplicit(mediapl)
					fmt.Println(mediapl.String())
					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
