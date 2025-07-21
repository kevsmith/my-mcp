//go:build darwin

package filesystem

import (
	"syscall"
	"time"
)

func extractCreationTime(stat interface{}) time.Time {
	if sysStat, ok := stat.(*syscall.Stat_t); ok {
		return time.Unix(sysStat.Birthtimespec.Sec, sysStat.Birthtimespec.Nsec)
	}
	return time.Time{}
}
