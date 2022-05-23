# Notes

- https://github.com/ItIckeYd/VodRecovery

```bash
go mod init goVods
git init
git branch -M main
touch main.go
go get github.com/samber/lo@v1
go get github.com/grafov/m3u8
```

## Todo

- [x] Get valid `index-dvr.m3u8` URLS for VOD
- [ ] Add option to filter out `unmuted.ts` files
- [ ] Add option to replace `unmuted.ts` files with `muted.ts` files

## Get the Streamer Information

We need three fields:

```go
type TwitchTrackerData struct {
	streamerName string
	videoId      int
	utcTime      string
}
```

- For example, we will consider `goonergooch`.
- We can go to https://twitchtracker.com/goonergooch/streams and find the relevant stream.
- Click on it.
  For example, the link might be https://twitchtracker.com/goonergooch/streams/46448856909.
- If we do `curl https://twitchtracker.com/goonergooch/streams/46448856909`, there will be a `div` element with the CSS classname `stream-timestamp-dt`.
  This gives the UTC time `2022-05-22 02:31:43`.
  Alternatively, we could just open `Chrome Developer Tools > Network > 46448856909 > Response` and search for that classname if you are using Google Chrome.
  Using `curl` and `grep` is easier though.

  ```bash
  curl --silent https://twitchtracker.com/goonergooch/streams/46448856909 | grep "stream-timestamp-dt"
  ```

  The time should be the first `div` entry.
  That corresponds to the stream start time.
  The second `div` entry corresponds to the stream end time.

- Our struct ends up being

  ```go
  {
    streamerName: "goonergooch",
    videoId: 46448856909,
    utcTime: "2022-05-22 02:31:43",
  }
  ```
