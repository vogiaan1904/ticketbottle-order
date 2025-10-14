package util

import (
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	DateTimeFormat = "2006-01-02 15:04:05"
	DateFormat     = "2006-01-02"
	TimeFormat     = "15:04"
	DDMMYYYYFormat = "02/01/2006"
)

func StrToDateTime(str string) (time.Time, error) {
	t, err := time.Parse(DateTimeFormat, str)
	if err != nil {
		return time.Time{}, err
	}

	// Create a new time.Time with the same date/time components but in local timezone
	localTz := GetDefaultTimezone()
	return time.Date(
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond(),
		localTz,
	), nil
}

func StrToDateTimeParse(str string) (time.Time, error) {
	t, err := time.Parse(DateTimeFormat, str)
	if err != nil {
		return time.Time{}, err
	}
	t = t.Add(-time.Duration(7) * time.Hour)
	return t.In(GetDefaultTimezone()), nil
}

func StrToDate(str string) (time.Time, error) {
	t, err := time.Parse(DateFormat, str)
	if err != nil {
		return time.Time{}, err
	}

	// Create a new time.Time with the same date/time components but in local timezone
	localTz := GetDefaultTimezone()
	return time.Date(
		t.Year(), t.Month(), t.Day(),
		0, 0, 0, 0,
		localTz,
	), nil
}

func DateTimeToStr(dt time.Time, ft *string) string {
	if ft == nil {
		return dt.Format(DateTimeFormat)
	} else {
		return dt.Format(*ft)
	}
}

func GetDefaultTimezone() *time.Location {
	localTimeZone, _ := time.LoadLocation("Local")
	return localTimeZone
}

func DateTimeToDDMMYYYY(dt time.Time) string {
	return dt.Format(DDMMYYYYFormat)
}

func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, GetDefaultTimezone())
}

func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, GetDefaultTimezone())
}

func Now() time.Time {
	return time.Now().In(GetDefaultTimezone())
}

func UnixToDateTime(unix int64) time.Time {
	return time.Unix(unix, 0).In(GetDefaultTimezone())
}

func GetPeriodAndYear(t time.Time) (int32, int32) {
	p := int32(math.Ceil(float64(t.Month()) / 3))
	return p, int32(t.Year())
}

func DaysInMonth(t time.Time) int {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()
}

func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func EndOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), DaysInMonth(t), 23, 59, 59, 999999999, t.Location())
}

func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

func EndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 23, 59, 59, 999999999, t.Location())
}

func GetPeriodAndYearRange(start_time time.Time, end_time time.Time) []time.Time {
	var result []time.Time
	var p_start_time, y_start_time int32
	var p_end_time, y_end_time int32

	// Start with the beginning of the first period
	start_time = StartOfMonth(start_time)
	result = append(result, start_time)

	for {
		p_start_time, y_start_time = GetPeriodAndYear(start_time)
		p_end_time, y_end_time = GetPeriodAndYear(end_time)

		if p_start_time == p_end_time && y_start_time == y_end_time {
			break
		}

		// Move to the start of the next period
		start_time = StartOfMonth(start_time.AddDate(0, 3, 0))
		result = append(result, start_time)

		if !start_time.Before(end_time) {
			break
		}
	}

	return result
}

func BuildYearMonthConditions(yearMonths []struct {
	Year  int
	Month int32
}) []bson.M {
	conditions := make([]bson.M, len(yearMonths))
	for i, ym := range yearMonths {
		conditions[i] = bson.M{
			"year":  ym.Year,
			"month": ym.Month,
		}
	}
	return conditions
}

// ConvertToLocalTimezone converts a time to the local timezone
func ConvertToLocalTimezone(eventTime time.Time, timezoneOffsetSeconds int) time.Time {
	loc := time.FixedZone("EventTimezone", timezoneOffsetSeconds)
	return eventTime.In(loc)
}
