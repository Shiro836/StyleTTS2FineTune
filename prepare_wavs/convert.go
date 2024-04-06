package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func convertToWav(ctx context.Context, dir string) {
	if _, err := os.Stat(path.Join(dir, convertLock)); err == nil {
		return
	}

	cmd := exec.Command("ffmpeg", "-i", path.Join(dir, downloaded), path.Join(dir, converted))

	errBuf := strings.Builder{}
	cmd.Stderr = &errBuf

	if err := cmd.Start(); err != nil {
		log.Fatal("start convert to wav:", err)
	}

	stop := make(chan struct{})

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Fatal("convert to wav:", err, errBuf.String())
		}
		close(stop)
	}()

	select {
	case <-stop:
	case <-ctx.Done():
		log.Fatal("interrupted")
	}

	if _, err := os.Create(path.Join(dir, convertLock)); err != nil {
		log.Fatal("create lock file:", err)
	}
}
