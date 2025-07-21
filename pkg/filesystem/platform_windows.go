//go:build windows

package filesystem

import (
	"syscall"
	"time"
)

func extractCreationTime(stat interface{}) time.Time {
	if winStat, ok := stat.(*syscall.Win32FileAttributeData); ok {
		return time.Unix(0, winStat.CreationTime.Nanoseconds())
	}
	return time.Time{}
}
