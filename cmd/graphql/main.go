package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/auoie/goVods/twitchgql"
)

type twitchStream struct {
	StreamId      string
	StreamerId    string
	StreamerLogin string
	ViewCount     int
	CreatedAt     time.Time
}

func toTwitchStream(edge twitchgql.GetStreamsStreamsStreamConnectionEdgesStreamEdge) *twitchStream {
	node := edge.Node
	broadcaster := node.Broadcaster
	return &twitchStream{
		StreamId:      node.Id,
		StreamerId:    broadcaster.Id,
		StreamerLogin: broadcaster.Login,
		ViewCount:     node.ViewersCount,
		CreatedAt:     node.CreatedAt,
	}
}

func main() {
	fmt.Println("Running...")
	httpClient := twitchgql.MakeTwitchqlClient()
	graphqlClient := graphql.NewClient("https://gql.twitch.tv/gql", httpClient)
	done := make(chan error)
	waiter := make(chan struct{})
	count := 0
	go func() {
		for {
			<-waiter
			time.Sleep(time.Millisecond * 500)
		}
	}()
	go func() {
		goForever := func() error {
			cursor := ""
			for {
				response, err := twitchgql.GetStreams(context.TODO(), graphqlClient, 30, cursor)
				if err != nil {
					return err
				}
				streams := response.Streams
				edges := streams.Edges
				if len(edges) == 0 {
					return err
				}
				twitchStreams := []*twitchStream{}
				for _, edge := range edges {
					twitchStreams = append(twitchStreams, toTwitchStream(edge))
				}
				bytes, err := json.MarshalIndent(twitchStreams, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(bytes))
				fmt.Println("Count:", count)
				fmt.Println(time.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00"))
				waiter <- struct{}{}
				count++
				lastCursor := edges[len(edges)-1].Cursor
				if lastCursor == "" {
					return errors.New("last cursor is empty")
				}
				cursor = lastCursor
			}
		}
		done <- goForever()
	}()
	err := <-done
	fmt.Print(err)
}
