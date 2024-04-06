package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func downloadYT(ctx context.Context, dir string, url string) {
	if _, err := os.Stat(path.Join(dir, dlLock)); err == nil {
		return
	}

	cmd := exec.Command("./youtube-dl", "-f", "bestaudio[ext=m4a]", "-o", path.Join(dir, downloaded), url)

	errBuf := strings.Builder{}
	cmd.Stderr = &errBuf

	if err := cmd.Start(); err != nil {
		log.Fatal("start denoise:", err)
	}

	stop := make(chan struct{})

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Fatal("download yt:", err, errBuf.String())
		}
		close(stop)
	}()

	select {
	case <-stop:
	case <-ctx.Done():
		log.Fatal("interrupted")
	}

	if _, err := os.Create(path.Join(dir, dlLock)); err != nil {
		log.Fatal("create lock file:", err)
	}
}
