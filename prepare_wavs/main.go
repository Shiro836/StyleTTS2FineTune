package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
)

const workDir = "inputs"
const datasets = "datasets"

func init() {
	os.Mkdir(workDir, os.ModePerm)
}

const dlLock = "dl_lock"
const convertLock = "convert_lock"
const denoiseLock = "denoise_lock"
const normLock = "norm_lock"
const whisperLock = "whisper_lock"
const cleanseLock = "cleanse_lock"
const splitLock = "split_lock"
const segment_lock = "segment_lock"

const downloaded = "audio"
const converted = "audio.wav"
const denoised = "audio_denoised.wav"
const normalized = "audio_denoised_normalized.wav"
const transcripted = "transcripted"
const cleansed = "cleansed.json"
const curated = "curate_this"
const segments = "segments"

func main() {
	var youtubeURL string

	var twitchVodURL string
	var clientID string

	var outputFolder string

	var dirs string

	flag.StringVar(&youtubeURL, "youtube_url", "", "youtube video url")

	flag.StringVar(&twitchVodURL, "twitch_url", "", "twitch vod/clip url")
	flag.StringVar(&clientID, "client_id", "kimne78kx3ncx6brgo4mv6wki5h1ko", "twitch client id")
	flag.StringVar(&outputFolder, "output_folder", "", "where to store processed audio")

	flag.StringVar(&dirs, "dirs", "", "dirs of speakers separated by comma")

	flag.Parse()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-c

		cancel()
	}()

	if len(dirs) != 0 {
		log.Println("crafting dataset")
		craftDataset(ctx, strings.Split(dirs, ","))

		return
	}

	if len(twitchVodURL) == 0 {
		if len(youtubeURL) == 0 {
			log.Fatal("twitch_url or youtube_url flag not provided")
		}
	}

	// TODO: make steps declarative or dependant on each other
	// TODO: make command as function with command arguments
	// TODO: use pipeline pattern to allow batched processing
	// TODO: use proxy servers to allow batched downloads
	// TODO: add start, end flag
	// TODO: use GPU deep-filter for faster denoising / split audio and parallelize current deep-filter

	dir := path.Join(workDir, outputFolder)

	if len(twitchVodURL) != 0 {
		log.Println("downloading twitch vod")
		download(ctx, dir, twitchVodURL, clientID)
	} else {
		log.Println("downloading youtube video")
		downloadYT(ctx, dir, youtubeURL)
	}

	log.Println("converting to wav")
	convertToWav(ctx, dir)

	log.Println("denoise audio")
	denoise(ctx, dir)

	log.Println("normalize audio")
	normalize(ctx, dir)

	log.Println("transcribe audio")
	whisper(ctx, dir)

	if _, err := os.Stat(path.Join(dir, "speakers.txt")); os.IsNotExist(err) {
		log.Println("speakers.txt file not found, creating it")
		os.WriteFile(path.Join(dir, "speakers.txt"), nil, 0644)
	}

	speakers, err := os.ReadFile(path.Join(dir, "speakers.txt"))
	if err != nil {
		log.Fatal("read speakers.txt file:", err)
	}
	if len(speakers) == 0 {
		log.Fatalf("add speakers separated by comma to speakers.txt, they are in %s folder in %s.srt file", path.Join(dir, transcripted), strings.TrimSuffix(normalized, ".wav"))
	}

	log.Println("cleansing transcript")
	cleanse(ctx, dir, string(speakers))

	log.Println("curating transcript")
	split(dir)

	log.Println("segmenting audio")
	segmentize(dir)
}
