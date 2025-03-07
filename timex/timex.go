// Package timex provides an enhanced time.Time implementation.
// Add more commonly used functional methods.
//
// such as: DayStart(), DayAfter(), DayAgo(), DateFormat() and more.
package timex

import (
	"time"

	"github.com/gookit/goutil/fmtutil"
	"github.com/gookit/goutil/strutil"
)

const (
	OneMinSec  = 60
	OneHourSec = 3600
	OneDaySec  = 86400
	OneWeekSec = 7 * 86400

	OneMin  = time.Minute
	OneHour = time.Hour
	OneDay  = 24 * time.Hour
	OneWeek = 7 * 24 * time.Hour
)

var (
	// DefaultLayout template for format time
	DefaultLayout = "2006-01-02 15:04:05"
)

// TimeX struct
type TimeX struct {
	time.Time
	// Layout set the default date format layout. default use DefaultLayout
	Layout string
}

// Now time
func Now() *TimeX {
	return &TimeX{
		Time:   time.Now(),
		Layout: DefaultLayout,
	}
}

// New form given time
func New(t time.Time) *TimeX {
	return &TimeX{
		Time:   t,
		Layout: DefaultLayout,
	}
}

// Local time for now
func Local() *TimeX {
	return New(time.Now().In(time.Local))
}

// FromUnix create from unix time
func FromUnix(sec int64) *TimeX {
	return New(time.Unix(sec, 0))
}

// FromString create from datetime string.
// see strutil.ToTime()
func FromString(s string, layouts ...string) (*TimeX, error) {
	t, err := strutil.ToTime(s, layouts...)
	if err != nil {
		return nil, err
	}

	return New(t), nil
}

// LocalByName time for now
func LocalByName(tzName string) *TimeX {
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		panic(err)
	}

	return New(time.Now().In(loc))
}

// SetLocalByName set local by tz name. eg: UTC, PRC
func SetLocalByName(tzName string) error {
	location, err := time.LoadLocation(tzName)
	if err != nil {
		return err
	}

	time.Local = location
	return nil
}

// Format returns a textual representation of the time value formatted according to the layout defined by the argument.
//
// see time.Time.Format()
func (t *TimeX) Format(layout string) string {
	if t.Layout == "" {
		layout = DefaultLayout
	}
	return t.Time.Format(layout)
}

// Datetime use DefaultLayout format time to date. see Format()
func (t *TimeX) Datetime() string {
	return t.Format(t.Layout)
}

// TplFormat use input template format time to date.
func (t *TimeX) TplFormat(template string) string {
	return t.DateFormat(template)
}

// DateFormat use input template format time to date.
// see ToLayout()
func (t *TimeX) DateFormat(template string) string {
	return t.Format(ToLayout(template))
}

// Yesterday get day ago time for the time
func (t *TimeX) Yesterday() *TimeX {
	return t.AddSeconds(-OneDaySec)
}

// DayAgo get some day ago time for the time
func (t *TimeX) DayAgo(day int) *TimeX {
	return t.AddSeconds(-day * OneDaySec)
}

// AddDay add some day time for the time
func (t *TimeX) AddDay(day int) *TimeX {
	return t.AddSeconds(day * OneDaySec)
}

// Tomorrow time. get tomorrow time for the time
func (t *TimeX) Tomorrow() *TimeX {
	return t.AddSeconds(OneDaySec)
}

// DayAfter get some day after time for the time.
// alias of TimeX.AddDay()
func (t *TimeX) DayAfter(day int) *TimeX {
	return t.AddDay(day)
}

// AddHour add some hour time
func (t *TimeX) AddHour(hours int) *TimeX {
	return t.AddSeconds(hours * OneHourSec)
}

// AddMinutes add some minutes time for the time
func (t *TimeX) AddMinutes(minutes int) *TimeX {
	return t.AddSeconds(minutes * OneMinSec)
}

// AddSeconds add some seconds time the time
func (t *TimeX) AddSeconds(seconds int) *TimeX {
	return &TimeX{
		Time: t.Add(time.Duration(seconds) * time.Second),
		// with layout
		Layout: DefaultLayout,
	}
}

// SubUnix calc diff seconds for t - u
func (t TimeX) SubUnix(u time.Time) int {
	return int(t.Sub(u) / time.Second)
}

// Diff calc diff duration for t - u.
// alias of time.Time.Sub()
func (t TimeX) Diff(u time.Time) time.Duration {
	return t.Sub(u)
}

// DiffSec calc diff seconds for t - u
func (t TimeX) DiffSec(u time.Time) int {
	return int(t.Sub(u) / time.Second)
}

// HourStart time
func (t *TimeX) HourStart() *TimeX {
	y, m, d := t.Date()
	newTime := time.Date(y, m, d, t.Hour(), 0, 0, 0, t.Location())

	return New(newTime)
}

// HourEnd time
func (t *TimeX) HourEnd() *TimeX {
	y, m, d := t.Date()
	newTime := time.Date(y, m, d, t.Hour(), 59, 59, int(time.Second-time.Nanosecond), t.Location())

	return New(newTime)
}

// DayStart time
func (t *TimeX) DayStart() *TimeX {
	y, m, d := t.Date()
	newTime := time.Date(y, m, d, 0, 0, 0, 0, t.Location())

	return New(newTime)
}

// DayEnd time
func (t *TimeX) DayEnd() *TimeX {
	y, m, d := t.Date()
	newTime := time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())

	return New(newTime)
}

// ChangeHMS change the hour, minute, second for create new time.
func (t *TimeX) ChangeHMS(hour, min, sec int) *TimeX {
	y, m, d := t.Date()
	newTime := time.Date(y, m, d, hour, min, sec, int(time.Second-time.Nanosecond), t.Location())

	return New(newTime)
}

// IsBefore the given time
func (t *TimeX) IsBefore(u time.Time) bool {
	return t.Before(u)
}

// IsAfter the given time
func (t *TimeX) IsAfter(u time.Time) bool {
	return t.After(u)
}

// HowLongAgo format diff time to string.
func (t TimeX) HowLongAgo(before time.Time) string {
	return fmtutil.HowLongAgo(t.Unix() - before.Unix())
}
