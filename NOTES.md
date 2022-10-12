# Notes

## Reuse HTTP Connections

I issue `60 * num_of_domains` HTTP requests in the case of the StreamsCharts data.
This is wasteful.
It takes like 10 seconds.
I should probably reuse HTTP connections.
See [here](https://husni.dev/how-to-reuse-http-connection-in-go/).
You can see what's going on with

```bash
time ./govods sc-manual-get-m3u8 --time "02-10-2022 01:31" --streamer goonergooch --videoid 47238989357 --write
netstat -atn
```

See [here](https://stackoverflow.com/questions/17948827/reusing-http-connections-in-go) for discussions and mistakes people made.

## HTTP/2 Multiplexing

It seems like there is something called HTTP/2 multiplexing.
You can issue multiple requests at the time time on a single TCP connection.
See [here](https://groups.google.com/g/golang-nuts/c/5T5aiDRl_cw).
They used Wireshark to assess this.

## Goroutines with Methods

See [here](https://stackoverflow.com/questions/36121984/how-to-use-a-method-as-a-goroutine-function).

## HTTP Transport

See [here](https://www.loginradius.com/blog/engineering/tune-the-go-http-client-for-high-performance/#problem2-default-http-transport)
for remarks on the Go default transport.

## GraphQL

[This repository](https://github.com/SuperSonicHub1/twitch-graphql-api) has the schema for the
Twitch GraphQL API.

```
curl -OL https://raw.githubusercontent.com/SuperSonicHub1/twitch-graphql-api/master/schema.graphql
```

See some basic usage, see [here](https://github.com/mauricew/twitch-graphql-api/blob/master/USAGE.md).

I'm not sure how to write GraphQL queries.
To get autocompletion, I installed [GraphQL: Language Feature Support in VSCode](https://marketplace.visualstudio.com/items?itemName=GraphQL.vscode-graphql).
It takes a while for the language server to process the schema, so you'll have to wait a bit if you reload VSCode.
If it's not working, look at the LSP logs in the VSCode output.
In order for it to work, I need a file `.graphqlrc.yml`.

```yaml
schema: "schema.graphql"
documents: "twitchgql/*.graphql"
```

Then we can use [Khan/genqlient](https://github.com/Khan/genqlient) to generate the GraphQL client code.
Following the pattern from [99designs/gqlgen](https://github.com/99designs/gqlgen#quick-start),
we create a file `tools.go` that contains `Khan/genqlient` as a dependency.

```bash
go mod tidy # Khan/genqlient stays because it is in tools.go
go run github.com/Khan/genqlient
```

Note order to see how to add a request header to every request for a client, see the `genqlient` [example](https://github.com/Khan/genqlient/blob/main/example/main.go).
There are more options discussed in [this StackOverflow post](https://stackoverflow.com/questions/54088660/add-headers-for-each-http-request-using-client).

Because of the annotation

```go
//go:generate go run github.com/Khan/genqlient genqlient.yaml
```

it is possible to just use `go generate ./...` to generate the client code.
See [here](https://eli.thegreenplace.net/2021/a-comprehensive-guide-to-go-generate/) for further details on code generation in Go.
