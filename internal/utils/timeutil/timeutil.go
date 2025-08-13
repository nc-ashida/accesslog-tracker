package timeutil

import (
	"errors"
	"time"
)

// FormatTimestamp は時間をRFC3339形式の文字列に変換します
func FormatTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// ParseTimestamp はRFC3339形式の文字列を時間に変換します
func ParseTimestamp(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, errors.New("empty timestamp string")
	}
	
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, err
	}
	
	return t, nil
}

// IsToday は指定された時間が今日かどうかを判定します
func IsToday(t time.Time) bool {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	target := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	
	return today.Equal(target)
}

// GetStartOfDay は指定された時間の日の開始時刻を返します
func GetStartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// GetEndOfDay は指定された時間の日の終了時刻を返します
func GetEndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// GetStartOfWeek は指定された時間の週の開始時刻を返します（月曜日開始）
func GetStartOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	} else {
		weekday--
	}
	
	return time.Date(t.Year(), t.Month(), t.Day()-int(weekday), 0, 0, 0, 0, t.Location())
}

// GetStartOfMonth は指定された時間の月の開始時刻を返します
func GetStartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// GetStartOfYear は指定された時間の年の開始時刻を返します
func GetStartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// IsWithinLastDays は指定された時間が過去N日以内かどうかを判定します
func IsWithinLastDays(t time.Time, days int) bool {
	now := time.Now()
	cutoff := now.AddDate(0, 0, -days)
	return t.After(cutoff)
}

// IsWithinLastHours は指定された時間が過去N時間以内かどうかを判定します
func IsWithinLastHours(t time.Time, hours int) bool {
	now := time.Now()
	cutoff := now.Add(-time.Duration(hours) * time.Hour)
	return t.After(cutoff)
}

// IsWithinLastMinutes は指定された時間が過去N分以内かどうかを判定します
func IsWithinLastMinutes(t time.Time, minutes int) bool {
	now := time.Now()
	cutoff := now.Add(-time.Duration(minutes) * time.Minute)
	return t.After(cutoff)
}
