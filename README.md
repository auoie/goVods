# README

## Usage

Build the binary.
If you don't want to build it, then just use `go run ./cmd/govods.go {stuff}`.

```bash
# Make the binary
go build ./cmd/govods.go
```

Run the program.
It should work with a `streamscharts.com` or `twitchtracker.com` url.
In case you get a 503 response, you can manually provide the stream data.
To manually provide the stream time, you can go directly to the link and inspect the response in Chrome DevTools.
In the case of `streamscharts.com`, the time field can be found in the `datetime` attribute of the first `<time>` element.

```bash
# Using a Streams Charts link, write the .m3u8 file to ./Downloads
./govods sc-url-get-m3u8 --write --url https://streamscharts.com/channels/{streamer}/streams/{videoid}

# Using a Streams Charts link, print the .m3u8 file to stdout
./govods sc-url-get-m3u8 --url https://streamscharts.com/channels/{streamer}/streams/{videoid}

# Using manually retrieved Stream Charts data, write the .m3u8 file to ./Downloads
./govods sc-manual-get-m3u8 --write --streamer {streamer} --videoid {videoid} --time {time}

# Using a Twitch Tracker link, write the .m3u8 file to ./Downloads
./govods tt-url-get-m3u8 --write --url https://twitchtracker.com/{streamer}/streams/{videoid}
```

Once we have the files, we can serve them over a local web server.

```bash
# Serve the files over a local web server
python3 -m http.server 8080 --directory Downloads
```

Then you can see the files in http://localhost:8080.
If you're using Google Chrome, you can install the extension
https://chrome.google.com/webstore/detail/native-hls-playback/emnphkkblegpebimobpbekeedfgemhof and then click on one of the files to play it.
Alternatively, you can use a media player such as MPV or VLC to play the files.

```bash
# Play a file with MPV
mpv http://localhost:8080/{streamername}/{stuff}.m3u8
```

## Using SullyGnome

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
- Then run the program with sullygnome.
  ```bash
  # Using manually retrieved Sully Gnome data, write the .m3u8 file to ./Downloads
  ./govods sg-manual-get-m3u8 --time {time} --streamer {streamer} --videoid {videoid} --write
  # Using manually retrieved Sully Gnome data, print the .m3u8 file to stdout
  ./govods sg-manual-get-m3u8 --time {time} --streamer {streamer} --videoid {videoid} --write
  ```

## About

- This is just used for downloading Twitch VODs that are sub only or unlisted.
  Basically just go to https://streamscharts.com/ and find the stream you want to download.
  Then copy that link and paste it into the program.
- If a VOD is public, there is an easier way get a VOD. You can get the URL of a HLS media manifest (stream download link)
  by going directly to the VOD and opening the Chrome Developer Tools > Network.
  It should be under Fetch/XHR.
  For example, the response might include

  ```text
  https://d1mhjrowxxagfy.cloudfront.net/{hash}_{username}_{videoId}_{time}/chunked/index-dvr.m3u8
  https://d1mhjrowxxagfy.cloudfront.net/{hash}_{username}_{videoId}_{time}/720p30/index-dvr.m3u8
  ```

  If there are muted segments, then the response might look like

  ```text
  https://d1ymi26ma8va5x.cloudfront.net/{hash}_{username}_{videoId}_{time}/chunked/index-muted-{unknown}.m3u8
  https://d1ymi26ma8va5x.cloudfront.net/{hash}_{username}_{videoId}_{time}/720p30/index-muted-{unknown}.m3u8
  ```

  Using one of the response urls, you can just do

  ```bash
  yt-dlp {url} --concurrent-fragments 4
  ```

  This is better because `yt-dlp` is a much more thoroughly tested piece of software.
  It is likely to be faster and have more features such as pausing the download.
  You could also use an application like MPV and VLC to view the video.

## Notes

- https://github.com/ItIckeYd/VodRecovery
- https://github.com/canhlinh/hlsdl
- https://github.com/melbahja/got
- https://github.com/yt-dlp/yt-dlp
- Some `.m3u8` files are much shorter than reported on twitchtracker.
  It seems that in this case, the `.m3u8` file is ending in

  ```
  #EXT-X-DISCONTINUITY
  #EXT-X-TWITCH-DISCONTINUITY
  #EXT-X-ENDLIST
  ```

  I was watching another stream and it ended with a stream warning disconnection.
  So it might happen when the stream goes down but starts up again.
  TwitchTracker reported the stream as a single stream, but the recording consisted of two separate VODs, each with their own video id. `streamscharts.com` seems to actually separate the two VODs.
  You can generally get the video id from there.

## Todo

- [x] Get valid `index-dvr.m3u8` URLS for VOD.
- [ ] Get the HSL master URL so that I can get all of the stream URLs.
      This will mean that I don't need to manually rewrite the `m3u8` segment names.
      Alternatively, find out how to get the unknown value in `index-muted-{unknown}.m3u8`.
