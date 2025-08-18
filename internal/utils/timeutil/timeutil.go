package timeutil

import (
	"errors"
	"fmt"
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
	daysToSubtract := int(weekday)
	if weekday == time.Sunday {
		daysToSubtract = 6 // 日曜日の場合、6日前の月曜日を取得
	} else {
		daysToSubtract = int(weekday) - 1 // 月曜日=1の場合、0日前、火曜日=2の場合、1日前
	}

	return time.Date(t.Year(), t.Month(), t.Day()-daysToSubtract, 0, 0, 0, 0, t.Location())
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

// ParseDate は日付文字列をパースします（YYYY-MM-DD形式）
func ParseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, errors.New("empty date string")
	}

	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}

// IsYesterday は指定された時間が昨日かどうかを判定します
func IsYesterday(t time.Time) bool {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	target := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	yesterdayDate := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())

	return target.Equal(yesterdayDate)
}

// IsThisWeek は指定された時間が今週かどうかを判定します
func IsThisWeek(t time.Time) bool {
	now := time.Now()
	weekStart := GetWeekStart(now)
	weekEnd := weekStart.AddDate(0, 0, 7)

	return t.After(weekStart) && t.Before(weekEnd)
}

// IsThisMonth は指定された時間が今月かどうかを判定します
func IsThisMonth(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month()
}

// GetWeekStart は指定された時間の週の開始時刻を返します（月曜日開始）
func GetWeekStart(t time.Time) time.Time {
	weekday := t.Weekday()
	daysToSubtract := int(weekday)
	if weekday == time.Sunday {
		daysToSubtract = 6 // 日曜日の場合、6日前の月曜日を取得
	} else {
		daysToSubtract = int(weekday) - 1 // 月曜日=1の場合、0日前、火曜日=2の場合、1日前
	}

	return time.Date(t.Year(), t.Month(), t.Day()-daysToSubtract, 0, 0, 0, 0, t.Location())
}

// GetMonthStart は指定された時間の月の開始時刻を返します
func GetMonthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// GetYearStart は指定された時間の年の開始時刻を返します
func GetYearStart(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// FormatDuration は時間間隔を文字列に変換します
func FormatDuration(d time.Duration) string {
	return d.String()
}

// FormatRelativeTime は相対時間を文字列に変換します
func FormatRelativeTime(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	} else if duration < 365*24*time.Hour {
		months := int(duration.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	} else {
		years := int(duration.Hours() / (24 * 365))
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}
