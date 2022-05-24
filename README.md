# README

## Usage

```bash
# Get .m3u8 files names
go run ./cmd/govods.go urls --url https://twitchtracker.com/{streamer}/streams/{video}

# Download the VOD to ./Downloads/{streamer}/{video}/index.ts
go run ./cmd/govods.go download --url https://twitchtracker.com/{streamer}/streams/{video}
```

## About

- This is just used for downloading Twitch VODs that are sub only or unlisted.
  Basically just go to https://twitchtracker.com/ and find the stream you want to download.
  Then copy that link and paste it into the program.
- If a VOD is public, then you can get the URL of a HLS media manifest (stream download link)
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

## Notes

- https://github.com/ItIckeYd/VodRecovery
- https://github.com/canhlinh/hlsdl
- https://github.com/melbahja/got
- https://github.com/yt-dlp/yt-dlp

## Todo

- [x] Get valid `index-dvr.m3u8` URLS for VOD.
- [ ] Add support for concurrent downloads
- [ ] Add the option to restart downloading if the download is interrupted.
- [ ] Get the HSL master URL so that I can get all of the stream URLs.
      This will mean that I don't need to write my own `m3u8` downloader.
      Alternatively, find out how to get the unknown value in `index-muted-{unknown}.m3u8`.
