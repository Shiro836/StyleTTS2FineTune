package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/iFaceless/godub"
)

const targetLen = 10 * time.Second

const curateMsg = "Delete all bad audios from /%s folder and change parameter below to True\nI_CURATED_FOLDER=False"

const curatedFile = "curated.txt"

func extractNum(str string) int {
	numStr := ""

	for _, c := range str {
		if c >= '0' && c <= '9' {
			numStr += string(c)
		}
	}

	num, err := strconv.Atoi(numStr)
	if err != nil {
		log.Fatal("failed to convert to int:", err)
	}

	return num
}

func segmentize(dir string) {
	if _, err := os.Stat(path.Join(dir, segment_lock)); err == nil {
		return
	}

	curatedMsg := fmt.Sprintf(curateMsg, curated)

	if _, err := os.Stat(path.Join(dir, curatedFile)); os.IsNotExist(err) {
		// log.Printf("%s file not found, creating it", curatedFile)
		os.WriteFile(path.Join(dir, curatedFile), []byte(curatedMsg), 0644)
	}

	curatedContent, err := os.ReadFile(path.Join(dir, curatedFile))
	if err != nil {
		log.Fatal("failed to read curated file:", err)
	}

	if string(curatedContent) == curatedMsg {
		log.Fatalf("please review audio files in /%s folder and delete some of them if they contain background noise/ have other speakers talking/ low quality or fix transcriptions, and then set I_CURATED_FOLDER=True in %s", curated, curatedFile)
	}

	os.Mkdir(path.Join(dir, segments), os.ModePerm)

	audio, _ := godub.NewLoader().Load(path.Join(dir, normalized))
	if err != nil {
		log.Fatal("failed to unmarshal normalized file:", err)
	}

	silence, err := godub.NewSilentAudioSegment(48000, audio.FrameRate())
	if err != nil {
		log.Fatal("failed to create silent audio segment:", err)
	}

	dirEntry, err := os.ReadDir(path.Join(dir, curated))
	if err != nil {
		log.Fatal("failed to read dir:", err)
	}

	audioFiles := []string{}
	transcriptionFiles := []string{}

	for _, entry := range dirEntry {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), ".wav") {
			audioFiles = append(audioFiles, entry.Name())
		} else if strings.HasSuffix(entry.Name(), ".txt") {
			transcriptionFiles = append(transcriptionFiles, entry.Name())
		}
	}

	sort.Slice(audioFiles, func(i, j int) bool {
		return extractNum(audioFiles[i]) < extractNum(audioFiles[j])
	})

	segmentText := ""
	segmentDur := time.Duration(0)

	var nextSegment *godub.AudioSegment

	counter := 0

	for _, audio := range audioFiles {
		num := extractNum(audio)
		transFile := fmt.Sprintf(transcriptionFile, num)
		if !slices.Contains(transcriptionFiles, transFile) {
			continue
		}

		transcriptionData, err := os.ReadFile(path.Join(dir, curated, transFile))
		if err != nil {
			log.Fatal("failed to read transcription file:", err)
		}
		transcription := string(transcriptionData)

		if len(transcription) <= 4 {
			continue
		}

		segment, err := godub.NewLoader().Load(path.Join(dir, curated, audio))
		if err != nil {
			log.Fatal("failed to unmarshal audio file:", err)
		}

		if len(segmentText) == 0 {
			segmentText = transcription
			segmentDur = segment.Duration()

			nextSegment = segment
		} else {
			if segmentText[len(segmentText)-1] != ' ' && transcription[0] != ' ' {
				segmentText += " "
			}
			segmentText += transcription
			segmentDur += segment.Duration()

			nextSegment, err = nextSegment.Append(segment)
			if err != nil {
				log.Fatal("failed to append audio:", err)
			}
		}

		segmentDur += 100 * time.Millisecond

		slicedSilence, err := silence.Slice(0, 100*time.Millisecond)
		if err != nil {
			log.Fatal("failed to slice audio:", err)
		}

		nextSegment, err = nextSegment.Append(slicedSilence)
		if err != nil {
			log.Fatal("failed to append audio:", err)
		}

		if segmentDur >= targetLen {
			err = godub.NewExporter(path.Join(dir, segments, fmt.Sprintf(audioFile, counter))).WithDstFormat("wav").Export(nextSegment)
			if err != nil {
				log.Fatal("failed to export audio:", err)
			}

			err = os.WriteFile(path.Join(dir, segments, fmt.Sprintf(transcriptionFile, counter)), []byte(segmentText), 0644)
			if err != nil {
				log.Fatal("failed to write transcription file:", err)
			}

			segmentText = ""
			counter++
		}
	}

	if _, err := os.Create(path.Join(dir, segment_lock)); err != nil {
		log.Fatal("create lock file:", err)
	}
}
