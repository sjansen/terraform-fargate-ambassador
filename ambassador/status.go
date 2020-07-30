package main

import (
	"context"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
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
		logger.Errorw("Unable to enumerate partitions.",
			"error", err,
		)
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
			logger.Errorw(p.Mountpoint,
				"type", p.Fstype,
				"error", err,
			)
		} else {
			logger.Infow(p.Mountpoint,
				"type", p.Fstype,
				"pct", humanize.FtoaWithDigits(u.UsedPercent, 1)+"%",
				"used", humanize.IBytes(u.Used),
				"total", humanize.IBytes(u.Total),
				"ipct", humanize.FtoaWithDigits(u.InodesUsedPercent, 1)+"%",
				"iused", humanize.SIWithDigits(float64(u.InodesUsed), 0, ""),
				"itotal", humanize.SIWithDigits(float64(u.InodesTotal), 0, ""),
			)
		}
	}
}
