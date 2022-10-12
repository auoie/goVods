package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Khan/genqlient/graphql"
	"github.com/auoie/goVods/twitchgql"
)

func main() {
	fmt.Println("Running...")
	httpClient := twitchgql.MakeTwitchqlClient()
	graphqlClient := graphql.NewClient("https://gql.twitch.tv/gql", httpClient)
	response, err := twitchgql.GetUserStream(context.TODO(), graphqlClient, "xqc")
	if err != nil {
		log.Fatal(err)
	}
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatal("failed to parse data")
	}
	fmt.Println(string(data))
}
