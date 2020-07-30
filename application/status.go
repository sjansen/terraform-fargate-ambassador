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

	reportDiskUsage(partitions)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Minute):
			reportDiskUsage(partitions)
		}
	}
}

func reportDiskUsage(partitions []disk.PartitionStat) {
	for _, p := range partitions {
		if u, err := disk.Usage(p.Mountpoint); err != nil {
			log.Err(err).
				Str("type", p.Fstype).
				Msg(p.Mountpoint)
		} else {
			log.Info().
				Str("type", p.Fstype).
				Str("pct", humanize.FtoaWithDigits(u.UsedPercent, 1)+"%").
				Str("used", humanize.IBytes(u.Used)).
				Str("total", humanize.IBytes(u.Total)).
				Str("ipct", humanize.FtoaWithDigits(u.InodesUsedPercent, 1)+"%").
				Str("iused", humanize.SIWithDigits(float64(u.InodesUsed), 0, "")).
				Str("itotal", humanize.SIWithDigits(float64(u.InodesTotal), 0, "")).
				Msg(p.Mountpoint)
		}
	}
}
