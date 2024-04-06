package main

import (
	"context"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"time"

	twitchdl "github.com/jybp/twitch-downloader"
)

func download(ctx context.Context, dir string, twitchVodURL string, clientID string) {
	if _, err := os.Stat(path.Join(dir, dlLock)); err == nil {
		return
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	reader, err := twitchdl.Download(ctx, client, clientID, twitchVodURL, "Audio Only", time.Duration(0), time.Duration(math.MaxInt64))
	if err != nil {
		log.Fatal("init download vod:", err)
	}

	os.Mkdir(dir, os.ModePerm)

	file, err := os.Create(path.Join(dir, downloaded))
	if err != nil {
		log.Fatal("create input file:", err)
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		log.Fatal("download vod:", err)
	}

	if _, err := os.Create(path.Join(dir, dlLock)); err != nil {
		log.Fatal("create lock file:", err)
	}
}
