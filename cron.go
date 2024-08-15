package hardloop

import (
	"time"

	"github.com/robfig/cron/v3"
)

// YearLimit is the maximum number of years to search for a matching time.
var YearLimit = 5

type Schedule interface {
	// Next returns the next time this schedule is activated, greater than the given time.
	Next(time.Time) time.Time
	// Prev returns the previous time this schedule is activated, less than the given time.
	Prev(time.Time) time.Time
}

type DelaySchedule interface {
	GetDelay() time.Duration
}

type SpecSchedule struct {
	ConstantDelaySchedule time.Duration

	*cron.SpecSchedule
}

var _ Schedule = &SpecSchedule{}

const (
	// Set the top bit if a star was included in the expression.
	starBit = 1 << 63
)

// Next returns the next time this schedule is activated, greater than the given
// time.  If no time can be found to satisfy the schedule, return the zero time.
func (s *SpecSchedule) Next(t time.Time) time.Time {
	if s.ConstantDelaySchedule != 0 {
		return t.Add(s.ConstantDelaySchedule - time.Duration(t.Nanosecond())*time.Nanosecond)
	}

	// General approach
	//
	// For Month, Day, Hour, Minute, Second:
	// Check if the time value matches.  If yes, continue to the next field.
	// If the field doesn't match the schedule, then increment the field until it matches.
	// While incrementing the field, a wrap-around brings it back to the beginning
	// of the field list (since it is necessary to re-verify previous field
	// values)

	// Convert the given time into the schedule's timezone, if one is specified.
	// Save the original timezone so we can convert back after we find a time.
	// Note that schedules without a time zone specified (time.Local) are treated
	// as local to the time provided.
	origLocation := t.Location()
	loc := s.Location
	if loc == time.Local {
		loc = t.Location()
	}
	if s.Location != time.Local {
		t = t.In(s.Location)
	}

	// Start at the earliest possible time (the upcoming second).
	t = t.Add(1*time.Second - time.Duration(t.Nanosecond())*time.Nanosecond)

	// This flag indicates whether a field has been incremented.
	added := false

	// If no time is found within five years, return zero.
	yearLimit := t.Year() + YearLimit

WRAP:
	if t.Year() > yearLimit {
		return time.Time{}
	}

	// Find the first applicable month.
	// If it's this month, then do nothing.
	for 1<<uint(t.Month())&s.Month == 0 {
		// If we have to add a month, reset the other parts to 0.
		if !added {
			added = true
			// Otherwise, set the date at the beginning (since the current time is irrelevant).
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
		}
		t = t.AddDate(0, 1, 0)

		// Wrapped around.
		if t.Month() == time.January {
			goto WRAP
		}
	}

	// Now get a day in that month.
	//
	// NOTE: This causes issues for daylight savings regimes where midnight does
	// not exist.  For example: Sao Paulo has DST that transforms midnight on
	// 11/3 into 1am. Handle that by noticing when the Hour ends up != 0.
	for !dayMatches(s, t) {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
		}
		t = t.AddDate(0, 0, 1)
		// Notice if the hour is no longer midnight due to DST.
		// Add an hour if it's 23, subtract an hour if it's 1.
		if t.Hour() != 0 {
			if t.Hour() > 12 {
				t = t.Add(time.Duration(24-t.Hour()) * time.Hour)
			} else {
				t = t.Add(time.Duration(-t.Hour()) * time.Hour)
			}
		}

		if t.Day() == 1 {
			goto WRAP
		}
	}

	for 1<<uint(t.Hour())&s.Hour == 0 {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, loc)
		}
		t = t.Add(1 * time.Hour)

		if t.Hour() == 0 {
			goto WRAP
		}
	}

	for 1<<uint(t.Minute())&s.Minute == 0 {
		if !added {
			added = true
			t = t.Truncate(time.Minute)
		}
		t = t.Add(1 * time.Minute)

		if t.Minute() == 0 {
			goto WRAP
		}
	}

	for 1<<uint(t.Second())&s.Second == 0 {
		if !added {
			added = true
			t = t.Truncate(time.Second)
		}
		t = t.Add(1 * time.Second)

		if t.Second() == 0 {
			goto WRAP
		}
	}

	return t.In(origLocation)
}

// Prev returns the prev time this schedule is activated, less than the given
// time.  If no time can be found to satisfy the schedule, return the zero time.
func (s *SpecSchedule) Prev(t time.Time) time.Time {
	if s.ConstantDelaySchedule != 0 {
		return t.Add(-1*s.ConstantDelaySchedule + time.Duration(t.Nanosecond())*time.Nanosecond)
	}
	// General approach
	//
	// For Month, Day, Hour, Minute, Second:
	// Check if the time value matches.  If yes, continue to the next field.
	// If the field doesn't match the schedule, then decriment the field until it matches.
	// While decrementing the field, a wrap-around brings it back to the beginning
	// of the field list (since it is necessary to re-verify previous field
	// values)

	// Convert the given time into the schedule's timezone, if one is specified.
	// Save the original timezone so we can convert back after we find a time.
	// Note that schedules without a time zone specified (time.Local) are treated
	// as local to the time provided.
	origLocation := t.Location()
	loc := s.Location
	if loc == time.Local {
		loc = t.Location()
	}
	if s.Location != time.Local {
		t = t.In(s.Location)
	}

	// Start at the earliest possible time (the upcoming second).
	t = t.Add(-1*time.Second + time.Duration(t.Nanosecond())*time.Nanosecond)

	// This flag indicates whether a field has been decremented.
	added := false

	// If no time is found within five years, return zero.
	yearLimit := t.Year() - YearLimit

WRAP:
	if t.Year() < yearLimit {
		return time.Time{}
	}

	// Find the first applicable month.
	// If it's this month, then do nothing.
	for 1<<uint(t.Month())&s.Month == 0 {
		// If we have to add a month, reset the other parts to 0.
		if !added {
			added = true
			// Otherwise, set the date at the beginning (since the current time is irrelevant).
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
		}

		t = t.AddDate(0, -1, 0)

		if 1<<uint(t.Month())&s.Month != 0 && t.Day() == 1 {
			year := t.Year()
			mount := t.Month()
			if t.Month() == time.December {
				year++
				mount = time.January
			} else {
				mount++
			}

			days := time.Date(year, mount, 0, 0, 0, 0, 0, loc).Day()

			t = t.AddDate(0, 0, days-1)
		}

		// Wrapped around.
		if t.Month() == time.December {
			goto WRAP
		}
	}

	// Now get a day in that month.
	for !dayMatches(s, t) {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
		}
		currentDay := t.Day()
		t = t.AddDate(0, 0, -1)

		if !dayMatches(s, t) && t.Hour() == 0 {
			t = t.Add(time.Hour * 23) //nolint:gomnd // 23 is the last hour of the day
		}

		if t.Day()-currentDay != 1 {
			goto WRAP
		}
	}

	for 1<<uint(t.Hour())&s.Hour == 0 {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, loc)
		}
		t = t.Add(-1 * time.Hour)

		if 1<<uint(t.Hour())&s.Hour != 0 && t.Minute() == 0 {
			t = t.Add(time.Minute * 59)
		}

		if t.Hour() == 0 {
			goto WRAP
		}
	}

	for 1<<uint(t.Minute())&s.Minute == 0 {
		if !added {
			added = true
			t = t.Truncate(time.Minute)
		}
		t = t.Add(-1 * time.Minute)

		if 1<<uint(t.Minute())&s.Minute != 0 && t.Second() == 0 {
			t = t.Add(time.Second * 59)
		}

		if t.Minute() == 0 {
			goto WRAP
		}
	}

	for 1<<uint(t.Second())&s.Second == 0 {
		if !added {
			added = true
			t = t.Truncate(time.Second)
		}
		t = t.Add(-1 * time.Second)

		// No need to check under seconds.

		if t.Second() == 0 {
			goto WRAP
		}
	}

	return t.In(origLocation)
}

// edited function of github.com/robfig/cron/v3 to changed day of week and day of month both usage.
func dayMatches(s *SpecSchedule, t time.Time) bool {
	var (
		domMatch bool = 1<<uint(t.Day())&s.Dom > 0
		dowMatch bool = 1<<uint(t.Weekday())&s.Dow > 0
	)

	if s.Dom&starBit > 0 && s.Dow&starBit > 0 {
		return domMatch || dowMatch
	}

	if !(s.Dom&starBit > 0) && !(s.Dow&starBit > 0) {
		return domMatch && dowMatch
	}

	if !(s.Dow&starBit > 0) && dowMatch {
		return dowMatch
	}

	if !(s.Dom&starBit > 0) && domMatch {
		return domMatch
	}

	return domMatch && dowMatch
}

// Parse returns a new cron schedule for the given spec.
//
// It accepts
//   - Standard crontab specs, e.g. "* * * * ?"
//   - Descriptors, e.g. "@midnight", "@every 1h30m"
func ParseStandard(spec string) (*SpecSchedule, error) {
	specSchedule, err := cron.ParseStandard(spec)
	if err != nil {
		return nil, err
	}

	if schedule, ok := specSchedule.(DelaySchedule); ok {
		return &SpecSchedule{ConstantDelaySchedule: schedule.GetDelay()}, nil
	}

	if schedule, ok := specSchedule.(cron.ConstantDelaySchedule); ok {
		return &SpecSchedule{ConstantDelaySchedule: schedule.Delay}, nil
	}

	return &SpecSchedule{SpecSchedule: specSchedule.(*cron.SpecSchedule)}, nil //nolint:forcetypeassert // no need to check
}

// Parser is default parser for cron to replacing the functions.
type Parser struct {
	ParseFn func(standardSpec string) (cron.Schedule, error)
}

// Parse returns a new cron schedule for the given spec.
// It use hardloop.Schedule interface instead of cron.Schedule.
func (p Parser) Parse(spec string) (cron.Schedule, error) { //nolint:ireturn // return interface to support other interfaces
	parseFn := p.ParseFn
	if parseFn == nil {
		parseFn = cron.ParseStandard
	}

	specSchedule, err := parseFn(spec)
	if err != nil {
		return nil, err
	}

	if schedule, ok := specSchedule.(DelaySchedule); ok {
		return &SpecSchedule{ConstantDelaySchedule: schedule.GetDelay()}, nil
	}

	if schedule, ok := specSchedule.(cron.ConstantDelaySchedule); ok {
		return &SpecSchedule{ConstantDelaySchedule: schedule.Delay}, nil
	}

	return &SpecSchedule{SpecSchedule: specSchedule.(*cron.SpecSchedule)}, nil //nolint:forcetypeassert // no need to check
}

// Parse2 is a helper function for parsing the spec and returning the SpecSchedule.
func (p Parser) Parse2(spec string) (Schedule, error) { //nolint:ireturn // return interface to support other interfaces
	v, err := p.Parse(spec)
	if err != nil {
		return nil, err
	}

	return v.(*SpecSchedule), nil //nolint:forcetypeassert // no need to check
}

var _ cron.ScheduleParser = &Parser{}
