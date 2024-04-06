package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func whisper(ctx context.Context, dir string) {
	if _, err := os.Stat(path.Join(dir, whisperLock)); err == nil {
		return
	}

	cmd := exec.Command("whisperx", path.Join(dir, normalized), "--output_dir", path.Join(dir, transcripted), "--diarize", "--language", "en", "--model", "large-v3", "--hf_token", "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", "--align_model", "WAV2VEC2_ASR_LARGE_LV60K_960H", "--initial_prompt", "Hey, I'm Michal from MadMonq and we're on our way to see Forsen. Hello, I am Forsen, the best gamer in the world.")
	// cmd.Env = append(cmd.Env, "CUDA_VISIBLE_DEVICES=1") // DOESN"T WORK!!!!!!!!!!!!!!! WTF!!~~!~

	errBuf := strings.Builder{}
	cmd.Stderr = &errBuf

	if err := cmd.Start(); err != nil {
		log.Fatal("start denoise:", err)
	}

	stop := make(chan struct{})

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Fatal("whisper:", err, errBuf.String())
		}
		close(stop)
	}()

	select {
	case <-stop:
	case <-ctx.Done():
		log.Fatal("interrupted")
	}

	if _, err := os.Create(path.Join(dir, whisperLock)); err != nil {
		log.Fatal("create lock file:", err)
	}
}
