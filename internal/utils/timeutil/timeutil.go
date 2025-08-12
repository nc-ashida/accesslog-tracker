package timeutil

import (
	"fmt"
	"strconv"
	"time"
)

// TimeZone はタイムゾーン定数を定義
const (
	UTC     = "UTC"
	JST     = "Asia/Tokyo"
	PST     = "America/Los_Angeles"
	EST     = "America/New_York"
	Default = UTC
)

// Format は時間フォーマット定数を定義
const (
	FormatRFC3339     = time.RFC3339
	FormatRFC3339Nano = time.RFC3339Nano
	FormatISO8601     = "2006-01-02T15:04:05Z07:00"
	FormatDate        = "2006-01-02"
	FormatTime        = "15:04:05"
	FormatDateTime    = "2006-01-02 15:04:05"
	FormatUnix        = "unix"
)

// Now は現在時刻をUTCで取得します
func Now() time.Time {
	return time.Now().UTC()
}

// NowInTimezone は指定されたタイムゾーンでの現在時刻を取得します
func NowInTimezone(timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone: %s", timezone)
	}
	return time.Now().In(loc), nil
}

// ParseTime は文字列を時間にパースします
func ParseTime(timeStr, format string) (time.Time, error) {
	switch format {
	case FormatUnix:
		unix, err := strconv.ParseInt(timeStr, 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid unix timestamp: %s", timeStr)
		}
		return time.Unix(unix, 0), nil
	default:
		return time.Parse(format, timeStr)
	}
}

// ToTimezone は時間を指定されたタイムゾーンに変換します
func ToTimezone(t time.Time, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone: %s", timezone)
	}
	return t.In(loc), nil
}

// StartOfDay は指定された日の開始時刻（00:00:00）を取得します
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay は指定された日の終了時刻（23:59:59.999999999）を取得します
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek は指定された週の開始日（月曜日）を取得します
func StartOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Monday {
		return StartOfDay(t)
	}
	daysToSubtract := int(weekday - time.Monday)
	return StartOfDay(t.AddDate(0, 0, -daysToSubtract))
}

// EndOfWeek は指定された週の終了日（日曜日）を取得します
func EndOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		return EndOfDay(t)
	}
	daysToAdd := int(time.Sunday - weekday)
	return EndOfDay(t.AddDate(0, 0, daysToAdd))
}

// StartOfMonth は指定された月の開始日を取得します
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth は指定された月の終了日を取得します
func EndOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month()+1, 0, 23, 59, 59, 999999999, t.Location())
}

// StartOfYear は指定された年の開始日を取得します
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear は指定された年の終了日を取得します
func EndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 23, 59, 59, 999999999, t.Location())
}

// AddDays は指定された日数を加算します
func AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

// AddHours は指定された時間数を加算します
func AddHours(t time.Time, hours int) time.Time {
	return t.Add(time.Duration(hours) * time.Hour)
}

// AddMinutes は指定された分数を加算します
func AddMinutes(t time.Time, minutes int) time.Time {
	return t.Add(time.Duration(minutes) * time.Minute)
}

// AddSeconds は指定された秒数を加算します
func AddSeconds(t time.Time, seconds int) time.Time {
	return t.Add(time.Duration(seconds) * time.Second)
}

// DurationBetween は2つの時間の間隔を取得します
func DurationBetween(start, end time.Time) time.Duration {
	return end.Sub(start)
}

// DaysBetween は2つの時間の間の日数を取得します
func DaysBetween(start, end time.Time) int {
	startDay := StartOfDay(start)
	endDay := StartOfDay(end)
	return int(endDay.Sub(startDay).Hours() / 24)
}

// HoursBetween は2つの時間の間の時間数を取得します
func HoursBetween(start, end time.Time) int {
	return int(end.Sub(start).Hours())
}

// MinutesBetween は2つの時間の間の分数を取得します
func MinutesBetween(start, end time.Time) int {
	return int(end.Sub(start).Minutes())
}

// IsToday は指定された時間が今日かどうかを判定します
func IsToday(t time.Time) bool {
	now := Now()
	return StartOfDay(t).Equal(StartOfDay(now))
}

// IsYesterday は指定された時間が昨日かどうかを判定します
func IsYesterday(t time.Time) bool {
	yesterday := AddDays(Now(), -1)
	return StartOfDay(t).Equal(StartOfDay(yesterday))
}

// IsThisWeek は指定された時間が今週かどうかを判定します
func IsThisWeek(t time.Time) bool {
	now := Now()
	return StartOfWeek(t).Equal(StartOfWeek(now))
}

// IsThisMonth は指定された時間が今月かどうかを判定します
func IsThisMonth(t time.Time) bool {
	now := Now()
	return StartOfMonth(t).Equal(StartOfMonth(now))
}

// IsThisYear は指定された時間が今年かどうかを判定します
func IsThisYear(t time.Time) bool {
	now := Now()
	return StartOfYear(t).Equal(StartOfYear(now))
}

// Age は指定された時間から現在までの経過時間を人間が読みやすい形式で返します
func Age(t time.Time) string {
	duration := DurationBetween(t, Now())
	
	years := int(duration.Hours() / 24 / 365)
	if years > 0 {
		if years == 1 {
			return "1年前"
		}
		return fmt.Sprintf("%d年前", years)
	}
	
	months := int(duration.Hours() / 24 / 30)
	if months > 0 {
		if months == 1 {
			return "1ヶ月前"
		}
		return fmt.Sprintf("%dヶ月前", months)
	}
	
	days := DaysBetween(t, Now())
	if days > 0 {
		if days == 1 {
			return "1日前"
		}
		return fmt.Sprintf("%d日前", days)
	}
	
	hours := HoursBetween(t, Now())
	if hours > 0 {
		if hours == 1 {
			return "1時間前"
		}
		return fmt.Sprintf("%d時間前", hours)
	}
	
	minutes := MinutesBetween(t, Now())
	if minutes > 0 {
		if minutes == 1 {
			return "1分前"
		}
		return fmt.Sprintf("%d分前", minutes)
	}
	
	return "今"
}

// ParseDuration は文字列を時間間隔にパースします
func ParseDuration(durationStr string) (time.Duration, error) {
	return time.ParseDuration(durationStr)
}

// FormatDuration は時間間隔を人間が読みやすい形式で返します
func FormatDuration(duration time.Duration) string {
	if duration < time.Minute {
		return fmt.Sprintf("%.0f秒", duration.Seconds())
	}
	
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		seconds := int(duration.Seconds()) % 60
		if seconds == 0 {
			return fmt.Sprintf("%d分", minutes)
		}
		return fmt.Sprintf("%d分%d秒", minutes, seconds)
	}
	
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		if minutes == 0 {
			return fmt.Sprintf("%d時間", hours)
		}
		return fmt.Sprintf("%d時間%d分", hours, minutes)
	}
	
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	if hours == 0 {
		return fmt.Sprintf("%d日", days)
	}
	return fmt.Sprintf("%d日%d時間", days, hours)
}
