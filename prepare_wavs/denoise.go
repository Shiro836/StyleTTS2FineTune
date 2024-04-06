package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func denoise(ctx context.Context, dir string) {
	if _, err := os.Stat(path.Join(dir, denoiseLock)); err == nil {
		return
	}

	// cmd := exec.Command("deepFilter", path.Join(dir, converted), "--output-dir", path.Join(dir, "out"), "--model-base-dir", "DeepFilterNet3", "--no-suffix", "--atten-lim", "100") // CUDA out of memory on long audios
	cmd := exec.Command("./deep-filter", path.Join(dir, converted), "--output-dir", path.Join(dir, "out"), "--model", "DeepFilterNet3_ll_onnx.tar.gz")

	errBuf := strings.Builder{}
	cmd.Stderr = &errBuf

	if err := cmd.Start(); err != nil {
		log.Fatal("start denoise:", err)
	}

	stop := make(chan struct{})

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Fatal("denoise:", err, errBuf.String())
		}
		close(stop)
	}()

	select {
	case <-stop:
	case <-ctx.Done():
		log.Fatal("interrupted")
	}

	if err := os.Rename(path.Join(dir, "out", converted), path.Join(dir, denoised)); err != nil {
		log.Fatal("rename denoised:", err)
	}

	if _, err := os.Create(path.Join(dir, denoiseLock)); err != nil {
		log.Fatal("create lock file:", err)
	}

	os.Remove(path.Join(dir, "out"))
}
