package hardloop

import (
	"fmt"
	"testing"
	"time"
)

func Test_ParseSchedule(t *testing.T) {
	type testTime struct {
		timeNow time.Time
		next    []time.Time
		prev    []time.Time
	}
	tests := []struct {
		message  string
		schedule string
		tests    []testTime
	}{
		{
			message:  "every day at 7:00",
			schedule: "0 7 * * *",
			tests: []testTime{
				{
					timeNow: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					next: []time.Time{
						time.Date(2023, 1, 1, 7, 0, 0, 0, time.UTC),
						time.Date(2023, 1, 2, 7, 0, 0, 0, time.UTC),
						time.Date(2023, 1, 3, 7, 0, 0, 0, time.UTC),
					},
					prev: []time.Time{
						time.Date(2022, 12, 31, 7, 0, 0, 0, time.UTC),
						time.Date(2022, 12, 30, 7, 0, 0, 0, time.UTC),
						time.Date(2022, 12, 29, 7, 0, 0, 0, time.UTC),
					},
				},
			},
		},
		{
			message:  "every 5 minutes",
			schedule: "*/5 * * * *",
			tests: []testTime{
				{
					timeNow: time.Date(2023, 1, 1, 1, 0, 0, 0, time.UTC),
					next: []time.Time{
						time.Date(2023, 1, 1, 1, 5, 0, 0, time.UTC),
						time.Date(2023, 1, 1, 1, 10, 0, 0, time.UTC),
						time.Date(2023, 1, 1, 1, 15, 0, 0, time.UTC),
						time.Date(2023, 1, 1, 1, 20, 0, 0, time.UTC),
					},
					prev: []time.Time{
						time.Date(2023, 1, 1, 0, 55, 0, 0, time.UTC),
						time.Date(2023, 1, 1, 0, 50, 0, 0, time.UTC),
						time.Date(2023, 1, 1, 0, 45, 0, 0, time.UTC),
					},
				},
			},
		},
		{
			message:  "every minutes",
			schedule: "* * * * *",
			tests: []testTime{
				{
					timeNow: time.Date(2023, 1, 1, 1, 0, 0, 0, time.UTC),
					next: []time.Time{
						time.Date(2023, 1, 1, 1, 1, 0, 0, time.UTC),
						time.Date(2023, 1, 1, 1, 2, 0, 0, time.UTC),
						time.Date(2023, 1, 1, 1, 3, 0, 0, time.UTC),
					},
					prev: []time.Time{
						time.Date(2023, 1, 1, 0, 59, 0, 0, time.UTC),
						time.Date(2023, 1, 1, 0, 58, 0, 0, time.UTC),
						time.Date(2023, 1, 1, 0, 57, 0, 0, time.UTC),
					},
				},
			},
		},
		{
			message:  "weekdays at 17:00",
			schedule: "0 17 * * 1,2,3,4,5",
			tests: []testTime{
				{
					timeNow: time.Date(2023, 1, 1, 5, 0, 0, 0, time.UTC),
					next: []time.Time{
						time.Date(2023, 1, 2, 17, 0, 0, 0, time.UTC),
						time.Date(2023, 1, 3, 17, 0, 0, 0, time.UTC),
						time.Date(2023, 1, 4, 17, 0, 0, 0, time.UTC),
						time.Date(2023, 1, 5, 17, 0, 0, 0, time.UTC),
						time.Date(2023, 1, 6, 17, 0, 0, 0, time.UTC),
						time.Date(2023, 1, 9, 17, 0, 0, 0, time.UTC),
						time.Date(2023, 1, 10, 17, 0, 0, 0, time.UTC),
					},
					prev: []time.Time{
						time.Date(2022, 12, 30, 17, 0, 0, 0, time.UTC),
						time.Date(2022, 12, 29, 17, 0, 0, 0, time.UTC),
						time.Date(2022, 12, 28, 17, 0, 0, 0, time.UTC),
						time.Date(2022, 12, 27, 17, 0, 0, 0, time.UTC),
						time.Date(2022, 12, 26, 17, 0, 0, 0, time.UTC),
						time.Date(2022, 12, 23, 17, 0, 0, 0, time.UTC),
					},
				},
			},
		},
		{
			message:  "every 5 minutes in specific days",
			schedule: "*/5 9-16 1-17,19-31 * 1-5",
			tests: []testTime{
				{
					timeNow: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					next: []time.Time{
						time.Date(2023, 1, 2, 9, 0, 0, 0, time.UTC),
						time.Date(2023, 1, 2, 9, 5, 0, 0, time.UTC),
					},
				},
				{
					// weekend will be monday
					timeNow: time.Date(2023, time.August, 5, 0, 0, 0, 0, time.UTC),
					next: []time.Time{
						time.Date(2023, time.August, 7, 9, 0, 0, 0, time.UTC),
					},
				},
				{
					// 18th day of month is not in schedule
					timeNow: time.Date(2023, time.August, 18, 0, 0, 0, 0, time.UTC),
					next: []time.Time{
						time.Date(2023, time.August, 21, 9, 0, 0, 0, time.UTC),
						time.Date(2023, time.August, 21, 9, 5, 0, 0, time.UTC),
						time.Date(2023, time.August, 21, 9, 10, 0, 0, time.UTC),
					},
					prev: []time.Time{
						time.Date(2023, time.August, 17, 16, 55, 0, 0, time.UTC),
					},
				},
				{
					// 18th day of month and it is weekend is not in schedule
					timeNow: time.Date(2023, time.November, 18, 0, 0, 0, 0, time.UTC),
					next: []time.Time{
						time.Date(2023, time.November, 20, 9, 0, 0, 0, time.UTC),
						time.Date(2023, time.November, 20, 9, 5, 0, 0, time.UTC),
						time.Date(2023, time.November, 20, 9, 10, 0, 0, time.UTC),
					},
					prev: []time.Time{
						time.Date(2023, time.November, 17, 16, 55, 0, 0, time.UTC),
					},
				},
			},
		},
		{
			message:  "to may",
			schedule: "0 17 23 5 *",
			tests: []testTime{
				{
					timeNow: time.Date(2020, time.January, 5, 10, 50, 0, 0, time.UTC),
					next: []time.Time{
						time.Date(2020, time.May, 23, 17, 0, 0, 0, time.UTC),
					},
					prev: []time.Time{
						time.Date(2019, time.May, 23, 17, 0, 0, 0, time.UTC),
					},
				},
			},
		},
		{
			message:  "year pass time",
			schedule: "35 17 * * 1,2,3,4,5",
			tests: []testTime{
				{
					timeNow: time.Date(2020, time.January, 5, 10, 20, 0, 0, time.UTC),
					prev: []time.Time{
						time.Date(2020, time.January, 3, 17, 35, 0, 0, time.UTC),
					},
				},
				{
					timeNow: time.Date(2020, time.January, 5, 10, 50, 0, 0, time.UTC),
					prev: []time.Time{
						time.Date(2020, time.January, 3, 17, 35, 0, 0, time.UTC),
					},
				},
			},
		},
		{
			message:  "random 1",
			schedule: "23 0-20/2 * * *",
			tests: []testTime{
				{
					timeNow: time.Date(2023, time.August, 5, 2, 26, 0, 0, time.UTC),
					next: []time.Time{
						time.Date(2023, time.August, 5, 4, 23, 0, 0, time.UTC),
					},
					prev: []time.Time{
						time.Date(2023, time.August, 5, 2, 23, 0, 0, time.UTC),
					},
				},
			},
		},
		{
			message:  "random 2",
			schedule: "0 0,12 1 */2 *",
			tests: []testTime{
				{
					timeNow: time.Date(2023, time.August, 5, 2, 26, 0, 0, time.UTC),
					next: []time.Time{
						time.Date(2023, time.September, 1, 0, 0, 0, 0, time.UTC),
					},
					prev: []time.Time{
						time.Date(2023, time.July, 1, 12, 0, 0, 0, time.UTC),
						time.Date(2023, time.July, 1, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
		{
			message:  "random 3",
			schedule: "0 0 1,15 * 3",
			tests: []testTime{
				{
					timeNow: time.Date(2023, time.August, 5, 2, 26, 0, 0, time.UTC),
					next: []time.Time{
						time.Date(2023, time.November, 1, 0, 0, 0, 0, time.UTC),
					},
					prev: []time.Time{
						time.Date(2023, time.March, 15, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("parse_%d", i), func(t *testing.T) {
			schedule, err := ParseStandard(tt.schedule)
			if err != nil {
				t.Fatalf("failed to parse schedule: %v", err)
			}

			for _, testN := range tt.tests {
				now := testN.timeNow

				for _, next := range testN.next {
					now = schedule.Next(now)
					if !now.Equal(next) {
						t.Fatalf("[next] %s expected %v, got %v, now %v", tt.schedule, next, now, testN.timeNow)
					}
				}

				now = testN.timeNow

				for _, prev := range testN.prev {
					now = schedule.Prev(now)
					if !now.Equal(prev) {
						t.Fatalf("[prev] %s expected %v, got %v, now %v", tt.schedule, prev, now, testN.timeNow)
					}
				}
			}
		})
	}
}
