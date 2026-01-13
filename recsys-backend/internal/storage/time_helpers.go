package storage

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func timeToDuration(t pgtype.Time) time.Duration {
	if !t.Valid {
		return 0
	}
	totalMicro := t.Microseconds
	return time.Duration(totalMicro) * time.Microsecond
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSeconds := int(d.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func minutesToDuration(minutes int) time.Duration {
	if minutes < 0 {
		return 0
	}
	return time.Duration(minutes) * time.Minute
}
