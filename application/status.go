package main

import (
	"context"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/disk"
)

type Status struct {
	Healthy       bool
	UptimeHuman   string
	UptimeSeconds uint
}

func GetStatus() Status {
	return Status{
		Healthy:       true,
		UptimeHuman:   humanize.Time(startTime),
		UptimeSeconds: uint(time.Since(startTime).Seconds()),
	}
}

func MonitorDiskUsage(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	partitions, err := disk.Partitions(false)
	if err != nil {
		log.Err(err).
			Msg("Unable to enumerate partitions.")
		return
	}

	for _, p := range partitions {
		log.Info().
			Str("dev", p.Device).
			Str("path", p.Mountpoint).
			Str("type", p.Fstype).
			Str("opts", p.Opts).
			Msg("Mountpoint Status")
	}

	reportDiskUsage([]string{"/"})
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Minute):
			reportDiskUsage([]string{"/"})
		}
	}
}

func reportDiskUsage(paths []string) {
	for _, p := range paths {
		if u, err := disk.Usage(p); err != nil {
			log.Err(err).Msg(p)
		} else {
			log.Info().
				Str("path", p).
				Uint64("used", u.Used).
				Uint64("total", u.Total).
				Str("pct", humanize.FtoaWithDigits(u.UsedPercent, 1)+"%").
				Str("hused", humanize.IBytes(u.Used)).
				Str("htotal", humanize.IBytes(u.Total)).
				Str("iused", humanize.SIWithDigits(float64(u.InodesUsed), 0, "")).
				Str("itotal", humanize.SIWithDigits(float64(u.InodesTotal), 0, "")).
				Str("ipct", humanize.FtoaWithDigits(u.InodesUsedPercent, 1)+"%").
				Msg("Disk Usage")
		}
	}
}
