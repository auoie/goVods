package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	govods "goVods"

	"github.com/grafov/m3u8"
	"github.com/samber/lo"
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

func downloadVod(videoData govods.VideoData) error {
	validPathIdentifiers, err := videoData.GetValidLinks(DOMAINS)
	if err != nil {
		return err
	}
	fmt.Println(strings.Join(lo.Map(validPathIdentifiers, func(val govods.DomainPathIdentifier, index int) string {
		return val.GetIndexDvrUrl()
	}), "\n"))
	if len(validPathIdentifiers) == 0 {
		return errors.New("no valid urls were found")
	}
	dpi := validPathIdentifiers[0]
	processedSegments, err := dpi.GetMediaSegments()
	if err != nil {
		return err
	}
	fmt.Println(lo.Map(processedSegments, func(val *m3u8.MediaSegment, index int) string { return val.URI }))
	directoryPath := filepath.Join("Downloads", videoData.StreamerName, fmt.Sprint(videoData.VideoId))
	if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
		return err
	}
	hlsdl := govods.NewHlsDl(dpi, directoryPath)
	err = hlsdl.DownloadSegments(processedSegments)
	if err != nil {
		return err
	}
	return nil
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
				Name:  "download",
				Usage: "Download a Twitch VOD",
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
					fmt.Println(twitchData)
					videoData, err := twitchData.GetVideoData()
					if err != nil {
						return err
					}
					return downloadVod(videoData)
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
