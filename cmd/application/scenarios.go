package main

import (
	"context"
	"crypto/rand"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

func FillDisk(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	content := make([]byte, 25*1024*1024)
	rand.Read(content)

	os.Mkdir("/tmp", 0777)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(15 * time.Second):
			if tmpfile, err := ioutil.TempFile("/tmp", "example"); err != nil {
				log.Err(err).Msg("Creating temporary file.")
			} else if _, err := tmpfile.Write(content); err != nil {
				log.Err(err).Msg("Filling temporary file.")
			} else if err := tmpfile.Sync(); err != nil {
				log.Err(err).Msg("Flushing temporary file.")
			} else if err := tmpfile.Close(); err != nil {
				log.Err(err).Msg("Closing temporary file.")
			}
		}
	}
}
