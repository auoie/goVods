# README

This CLI tool helps obtain the M3U8 file for Twitch VODs.
It works even if a VOD is sub-only or unlisted.
If the VOD is not on the Twitch servers, (i.e. one of their S3 buckets),
then it is impossible to recover a VOD.

## Building

Build the binary.
If you don't want to build it, then just use `go run ./cmd/govods.go {stuff}`.

```bash
go build ./cmd/govods.go # Make the binary
./govods --help # View help
./govods sg-manual-get-m3u8 --help # View help for command sg-manual-get-m3u8
```

## Fetch a Single VOD

### Using SullyGnome

- Open the network tab in developer tools
- Go to `https://sullygnome.com/channel/{streamer}/streams`
- There should be a response with the XHR response type. View it.
- This response should have information about the recent streams in JSON format. It should include the following fields.
  ```jsonc
  {
    "startDateTime": "time", // start time of the stream in 2006-01-02T15:04:05Z format
    "streamId": "videoid" // video id of the stream
  }
  ```
- Then run the program with SullyGnome.
  ```bash
  # Using manually retrieved SullyGnome data, write the .m3u8 file to ./Downloads
  ./govods sg-manual-get-m3u8 --time {time} --streamer {streamer} --videoid {videoid} --write
  # Using manually retrieved SullyGnome data, print the .m3u8 file to stdout
  ./govods sg-manual-get-m3u8 --time {time} --streamer {streamer} --videoid {videoid}
  ```

### Using StreamsCharts

- Go to `https://streamscharts.com/channels/{streamer}/streams` to find a streamer's recent streams.
  Then go to the stream you want at `https://streamscharts.com/channels/{streamer}/streams/{videoid}`.
- Open the Elements Tab in the Chrome developer tools.
- Select the first `<time>` element with the `datetime` attribute.
  Then copy the `datetime` attribute. It should be in the format `02-01-2006 15:04`.
  You can get this value in the console with
  ```javascript
  document.querySelector("time[datetime]").getAttribute("datetime");
  ```
- Then run the program with StreamsCharts.
  ```bash
  # Using manually retrieved StreamCharts data, write the .m3u8 file to ./Downloads
  ./govods sc-manual-get-m3u8 --streamer {streamer} --videoid {videoid} --time {time} --write
  # Using manually retrieved StreamsCharts data, print the .m3u8 file to stdout
  ./govods sc-manual-get-m3u8 --streamer {streamer} --videoid {videoid} --time {time}
  ```

### Using TwitchTracker

- Go to `https://twitchtracker.com/{streamer}/streams` for the streamer's streams.
- Open the Network tab in the Chrome developer tools.
  Then click on the stream you want.
- One of the responses should be the HTML. You can filter for it by clicking on the `Doc` filter near the top.
  Click on it. Then click on the response tab to show the actual HTML.
- Search for `Stream started` in the HTML. Right above it should be the start time in `2006-01-02 15:04:05` format.
- Then run the program with TwitchTracker.
  ```bash
  # Using manually retrieved TwitchTracker data, write the .m3u8 file to ./Downloads
  ./govods tt-manual-get-m3u8 --streamer {streamer} --videoid {videoid} --time {time} --write
  # Using manually retrieved TwitchTracker data, write the .m3u8 file to stdout
  ./govods tt-manual-get-m3u8 --streamer {streamer} --videoid {videoid} --time {time}
  ```

## Viewing or Downloading a VOD

Once we have fetched the files, we can serve them over a local web server.
For example, if you have `python3` installed locally, you can run

```bash
# Serve the files over a local web server
python3 -m http.server 8080 --directory Downloads
```

Then you can see the files in `http://localhost:8080`.
If you're using Google Chrome, you can install the [Native HLS Playback](https://chrome.google.com/webstore/detail/native-hls-playback/emnphkkblegpebimobpbekeedfgemhof) and then click on one of the files to play it.
Alternatively, you can use a media player such as MPV or VLC to play the files.

You can also download the VOD locally with `yt-dlp`.

```bash
# Play a file with MPV
mpv http://localhost:8080/{streamername}/{stuff}.m3u8
# Download the VOD
yt-dlp http://localhost:8080/{streamername}/{stuff}.m3u8 --concurrent-fragments 4
```

## Notes and References

- https://github.com/TwitchRecover/TwitchRecover
- https://github.com/ItIckeYd/VodRecovery
- https://github.com/canhlinh/hlsdl
- https://github.com/melbahja/got
- https://github.com/yt-dlp/yt-dlp
- Some `.m3u8` files are much shorter than reported on TwitchTracker.
  It seems that in this case, the `.m3u8` file is ending in

  ```
  #EXT-X-DISCONTINUITY
  #EXT-X-TWITCH-DISCONTINUITY
  #EXT-X-ENDLIST
  ```

  I was watching another stream, and it ended with a stream warning disconnection.
  So it might happen when the stream goes down but starts up again.
  TwitchTracker reported the stream as a single stream, but the recording consisted of two separate VODs, each with their own video id. `streamscharts.com` seems to be the only website that separates the two VODs.
