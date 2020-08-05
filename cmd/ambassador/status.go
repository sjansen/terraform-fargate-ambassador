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

	for _, p := range partitions {
		logger.Infow("Mountpoint Status",
			"path", p.Mountpoint,
			"dev", p.Device,
			"type", p.Fstype,
			"opts", p.Opts,
		)
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
			logger.Errorw(p, "error", err)
		} else {
			logger.Infow("Disk Usage",
				"path", p,
				"used", u.Used,
				"total", u.Total,
				"pct", humanize.FtoaWithDigits(u.UsedPercent, 1)+"%",
				"hused", humanize.IBytes(u.Used),
				"htotal", humanize.IBytes(u.Total),
				"iused", humanize.SIWithDigits(float64(u.InodesUsed), 0, ""),
				"itotal", humanize.SIWithDigits(float64(u.InodesTotal), 0, ""),
				"ipct", humanize.FtoaWithDigits(u.InodesUsedPercent, 1)+"%",
			)
		}
	}
}
