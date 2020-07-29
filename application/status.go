package main

import (
	"context"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/disk"
)

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
				Float64("pct", u.UsedPercent).
				Str("used", humanize.IBytes(u.Used)).
				Str("total", humanize.IBytes(u.Total)).
				Float64("ipct", u.InodesUsedPercent).
				Uint64("iused", u.InodesUsed).
				Uint64("itotal", u.InodesTotal).
				Msg(p.Mountpoint)
		}
	}
}
