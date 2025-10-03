package hardloop

import "time"

type ScheduleGroup struct {
	StartSchedules []Schedule
	StopSchedules  []Schedule
}

func NewSchedule(startSpec, endSpec []string) (*ScheduleGroup, error) {
	startSchedules := make([]Schedule, 0, len(startSpec))
	stopSchedules := make([]Schedule, 0, len(endSpec))

	for _, spec := range startSpec {
		startSchedule, err := ParseStandard(spec)
		if err != nil {
			return nil, err
		}

		startSchedules = append(startSchedules, startSchedule)
	}

	for _, spec := range endSpec {
		stopSchedule, err := ParseStandard(spec)
		if err != nil {
			return nil, err
		}

		stopSchedules = append(stopSchedules, stopSchedule)
	}

	return &ScheduleGroup{
		StartSchedules: startSchedules,
		StopSchedules:  stopSchedules,
	}, nil
}

// getStartTime if return nil, start now.
func (l *ScheduleGroup) getStartTime(now time.Time) (*time.Time, error) {
	nextStart := FindNext(l.StartSchedules, now)

	if nextStart.IsZero() {
		return nil, errTimeNotSet
	}

	return &nextStart, nil
}

// getStopTime if return nil, stop now.
func (l *ScheduleGroup) getStopTime(now time.Time) (*time.Time, error) {
	prevStop := FindPrev(l.StopSchedules, now)

	if prevStop.IsZero() {
		// stop the ScheduleGroup
		return nil, errTimeNotSet
	}

	prevStart := FindPrev(l.StartSchedules, now)

	// if prevStop is after prevStart, then we should stop the loop
	if !prevStart.IsZero() && prevStop.After(prevStart) {
		// stop the loop
		return nil, nil
	}

	nextStop := FindNext(l.StopSchedules, now)

	if nextStop.IsZero() {
		// stop the loop
		return nil, errTimeNotSet
	}

	return &nextStop, nil
}

// NextStartTime returns the next start time.
//   - If it should be start now than it returns nil.
func (l *ScheduleGroup) NextStartTime(now time.Time) *time.Time {
	stopTime, _ := l.getStopTime(now)
	if stopTime != nil {
		return nil
	}

	startTime, _ := l.getStartTime(now)
	return startTime
}
