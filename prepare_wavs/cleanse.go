package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

type overlap struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type pyannoteResult struct {
	Overlaps []overlap `json:"overlaps"`
}

type word struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type segment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`

	Text    string `json:"text"`
	Speaker string `json:"speaker"`

	Words []word `json:"words"`
}

type whisperxResult struct {
	Segments []segment `json:"segments"`
}

const silenceBuffer = 0.1 // seconds

func cleanse(ctx context.Context, dir string, speakersStr string) {
	if _, err := os.Stat(path.Join(dir, cleanseLock)); err == nil {
		return
	}

	cmd := exec.CommandContext(ctx, "python", "find_overlaps.py", path.Join(dir, normalized))

	errBuf := strings.Builder{}
	cmd.Stderr = &errBuf

	res := strings.Builder{}
	cmd.Stdout = &res

	if err := cmd.Run(); err != nil {
		log.Fatal("start denoise:", err)
	}

	ind := strings.Index(res.String(), "{\"overlap")
	if ind == -1 {
		log.Fatal("malformed result, out:", res.String(), "| err:", errBuf.String())
	}

	resStr := res.String()[ind:]

	var result pyannoteResult
	if err := json.Unmarshal([]byte(resStr), &result); err != nil {
		log.Fatal("parse result:", err)
	}

	whisperfile, err := os.ReadFile(path.Join(dir, transcripted, strings.TrimSuffix(normalized, filepath.Ext(normalized))+".json"))
	if err != nil {
		log.Fatal("read whisperx result:", err)
	}

	var whisperxResult whisperxResult
	if err := json.Unmarshal(whisperfile, &whisperxResult); err != nil {
		log.Fatal("parse whisperx result:", err)
	}

	speakers := strings.Split(speakersStr, ",")

	total := len(whisperxResult.Segments)
	skppedNonSpeakers := 0
	skippedOverlaps := 0

	whisperxResult.Segments = slices.DeleteFunc(whisperxResult.Segments, func(s segment) bool {
		if !slices.Contains(speakers, s.Speaker) {
			skppedNonSpeakers++
			return true
		}

		for _, overlap := range result.Overlaps {
			if overlap.Start > s.Start && overlap.Start < s.End {
				skippedOverlaps++
				return true
			}
			if overlap.End > s.Start && overlap.End < s.End {
				skippedOverlaps++
				return true
			}
		}

		for _, segment := range whisperxResult.Segments {
			if segment.Start == s.Start && segment.End == s.End {
				continue
			}

			if segment.Start > s.Start && segment.Start < s.End {
				skippedOverlaps++
				return true
			}
			if segment.End > s.Start && segment.End < s.End {
				skippedOverlaps++
				return true
			}
		}

		for _, segment := range whisperxResult.Segments {
			if segment.End < s.Start && segment.End > s.Start-silenceBuffer {
				skippedOverlaps++
				return true
			}
			if segment.Start > s.End && segment.Start < s.End+silenceBuffer {
				skippedOverlaps++
				return true
			}
		}

		return false
	})

	log.Printf("removed %d non targeted speakers and %d overlaps out of %d total", skppedNonSpeakers, skippedOverlaps, total)

	resultFile, err := json.MarshalIndent(whisperxResult, "", "  ")
	if err != nil {
		log.Fatal("pretty print result:", err)
	}

	if err := os.WriteFile(path.Join(dir, cleansed), resultFile, 0644); err != nil {
		log.Fatal("write cleanse result:", err)
	}

	if _, err := os.Create(path.Join(dir, cleanseLock)); err != nil {
		log.Fatal("create lock file:", err)
	}
}
