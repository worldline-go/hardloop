package hardloop_test

import (
	"testing"
	"time"

	"github.com/worldline-go/hardloop"
)

func TestLoop_NextStartTime(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		startSpec []string
		endSpec   []string
		// Named input parameters for target function.
		now  time.Time
		want *time.Time
	}{
		{
			name: "out of range before start",
			startSpec: []string{
				"0 12 * * *", // at 12
			},
			endSpec: []string{
				"0 17 * * *", // at 17
			},
			now: time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
			want: func() *time.Time {
				t := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
				return &t
			}(),
		},
		{
			name: "same time",
			startSpec: []string{
				"0 12 * * *", // at 12
			},
			endSpec: []string{
				"0 17 * * *", // at 17
			},
			now:  time.Date(2024, 1, 2, 13, 0, 0, 0, time.UTC),
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := hardloop.NewSchedule(tt.startSpec, tt.endSpec)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got := l.NextStartTime(tt.now)
			if (got == nil) != (tt.want == nil) || (got != nil && !got.Equal(*tt.want)) {
				t.Errorf("NextStartTime() got = %v, want %v", got, tt.want)
			}
		})
	}
}
