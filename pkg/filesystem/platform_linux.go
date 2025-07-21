//go:build linux

package filesystem

import (
	"syscall"
	"time"
)

func extractCreationTime(stat interface{}) time.Time {
	if sysStat, ok := stat.(*syscall.Stat_t); ok {
		// Linux doesn't have birth time, use ctime (status change time) as fallback
		return time.Unix(sysStat.Ctim.Sec, sysStat.Ctim.Nsec)
	}
	return time.Time{}
}
