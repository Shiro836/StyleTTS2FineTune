package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"
)

const valPercentage = 0.08

func craftDataset(ctx context.Context, dirs []string) {
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i] < dirs[j]
	})

	os.Mkdir(datasets, os.ModePerm)

	finalDir := ""
	for i, dir := range dirs {
		if i != 0 {
			finalDir += "_"
		}
		finalDir += dir
	}

	finalDir = path.Join(datasets, finalDir)
	os.Mkdir(finalDir, os.ModePerm)
	os.Mkdir(path.Join(finalDir, "wavs"), os.ModePerm)

	for i := range dirs {
		dirs[i] = path.Join(workDir, dirs[i])
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			log.Fatalf("directory %s does not exist", dir)
		}

		if _, err := os.Stat(path.Join(dir, segments)); os.IsNotExist(err) {
			log.Fatalf("directory %s does not exist", path.Join(dir, segments))
		}

		dirEntry, err := os.ReadDir(path.Join(dir, segments))
		if err != nil {
			log.Fatalf("error reading directory %s", path.Join(dir, segments))
		}
		if len(dirEntry) == 0 {
			log.Fatalf("directory %s is empty", path.Join(dir, segments))
		}
	}

	counter := 0

	// csv | delimiter = "|", newline = "\n"
	trainFile := strings.Builder{}
	valFile := strings.Builder{}

	for _, dir := range dirs {
		files, err := os.ReadDir(path.Join(dir, segments))
		if err != nil {
			log.Fatalf("error reading directory %s", path.Join(dir, segments))
		}

		audios := []string{}

		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".wav") {
				audios = append(audios, file.Name())
			}
		}

		sort.Slice(audios, func(i, j int) bool {
			return extractNum(audios[i]) < extractNum(audios[j])
		})

		valCntFloat := float64(len(audios)) * valPercentage
		valCnt := max(1, int(valCntFloat))

		dirCounter := 0

		for _, audio := range audios {
			num := extractNum(audio)
			transPath := path.Join(dir, segments, fmt.Sprintf(transcriptionFile, num))
			transcription, err := os.ReadFile(transPath)
			if err != nil {
				log.Fatalf("error reading file %s", transPath)
			}

			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			nextFileName := fmt.Sprintf(audioFile, counter)

			cmd := exec.CommandContext(ctx, "ffmpeg", "-i", path.Join(dir, segments, audio), "-acodec", "pcm_s16le", "-ar", "24000", "-ac", "1", path.Join(finalDir, "wavs", nextFileName))
			err = cmd.Run()
			if err != nil {
				log.Fatalf("error running ffmpeg %s", cmd.String())
			}

			cancel()

			var fileToWrite *strings.Builder = &trainFile
			if dirCounter < valCnt {
				fileToWrite = &valFile
			}

			_, err = fileToWrite.WriteString(fmt.Sprintf("%s|%s\n", nextFileName, strings.TrimSpace(string(transcription))))
			if err != nil {
				log.Fatalf("error writing to final trans file")
			}

			dirCounter++
			counter++
		}
	}

	err := os.WriteFile(path.Join(finalDir, "train_list.txt"), []byte(trainFile.String()), os.ModePerm)
	if err != nil {
		log.Fatalf("error writing to final trans file")
	}
	err = os.WriteFile(path.Join(finalDir, "val_list.txt"), []byte(valFile.String()), os.ModePerm)
	if err != nil {
		log.Fatalf("error writing to final val file")
	}
}
