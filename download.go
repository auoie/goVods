package govods

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func downloadVideo(filepath string, url string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	fmt.Println("Downloading to", filepath)
	_, err = io.Copy(out, res.Body)
	return err
}

func DownloadTwitchVod(videoData TwitchTrackerData, dpi DomainPathIdentifier, numWorkers int) {
	directory := filepath.Join("Downloads", videoData.StreamerName, fmt.Sprint(videoData.VideoId))
	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	processedUris, err := getChunkUris(dpi)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(processedUris)
	jobs := []func() error{}
	for _, processedUri := range processedUris {
		url := dpi.getChunkUrl(processedUri)
		filePath := filepath.Join(directory, processedUri)
		job := func() error {
			return downloadVideo(filePath, url)
		}
		jobs = append(jobs, job)
	}
	for _, job := range jobs {
		job()
	}
}
