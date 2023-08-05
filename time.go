package hardloop

import "time"

func FindPrev(schedules []Schedule, now time.Time) time.Time {
	prevTime := time.Time{}

	for _, schedule := range schedules {
		prev := schedule.Prev(now)
		if prev.IsZero() {
			continue
		}

		if prevTime.IsZero() {
			prevTime = prev

			continue
		}

		if prev.After(prevTime) {
			prevTime = prev
		}
	}

	return prevTime
}

func FindNext(schedules []Schedule, now time.Time) time.Time {
	nextTime := time.Time{}
	for _, schedule := range schedules {
		next := schedule.Next(now)
		if next.IsZero() {
			continue
		}

		if nextTime.IsZero() {
			nextTime = next

			continue
		}

		if next.Before(nextTime) {
			nextTime = next
		}
	}

	return nextTime
}
