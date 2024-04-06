package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/iFaceless/godub"
)

const audioFile = "audio_%d_file.wav"
const transcriptionFile = "audio_%d_transcription.txt"

func split(dir string) {
	if _, err := os.Stat(path.Join(dir, splitLock)); err == nil {
		return
	}

	os.Mkdir(path.Join(dir, curated), os.ModePerm)

	cleansedFile, err := os.ReadFile(path.Join(dir, cleansed))
	if err != nil {
		log.Fatal("failed to read cleansed file:", err)
	}

	var cleansed whisperxResult
	if err := json.Unmarshal(cleansedFile, &cleansed); err != nil {
		log.Fatal("failed to unmarshal whisperx result:", err)
	}

	audio, _ := godub.NewLoader().Load(path.Join(dir, normalized))
	if err != nil {
		log.Fatal("failed to unmarshal normalized file:", err)
	}

	counter := 0

	for _, segment := range cleansed.Segments {
		segment.Text = strings.ReplaceAll(segment.Text, "...", ".")
		segment.Text = strings.ReplaceAll(segment.Text, ". . .", ".")

		if len(segment.Text) <= 1 {
			continue
		}

		audioSegment, err := audio.Slice(max(0, time.Duration(segment.Start*float64(time.Second))-time.Millisecond*90), time.Duration(segment.End*float64(time.Second))+time.Millisecond*90)
		if err != nil {
			log.Fatal("failed to slice audio segment:", err)
		}

		err = godub.NewExporter(path.Join(dir, curated, fmt.Sprintf(audioFile, counter))).WithDstFormat("wav").Export(audioSegment)
		if err != nil {
			log.Fatal("failed to export audio segment:", err)
		}

		err = os.WriteFile(path.Join(dir, curated, fmt.Sprintf(transcriptionFile, counter)), []byte(segment.Text), os.ModePerm)
		if err != nil {
			log.Fatal("failed to write transcription file:", err)
		}

		counter++
	}

	if _, err := os.Create(path.Join(dir, splitLock)); err != nil {
		log.Fatal("create lock file:", err)
	}
}
