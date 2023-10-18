package fetcher

import (
	"testing"
	"time"
)

func calculateRelativeTS(now time.Time, ts []time.Time) []time.Duration {
	rel := make([]time.Duration, len(ts))
	for i, ts := range ts {
		rel[i] = now.Sub(ts)
	}
	return rel
}

func Test_pruneCallsLocked(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		timestamps []time.Time
		ts         time.Time
		inspect    func(*testing.T, []time.Time)
	}{
		{
			"no previous calls",
			nil,
			now,
			func(t *testing.T, ts []time.Time) {
				if len(ts) != 0 {
					t.Errorf("unexpected number of timestamps want=0, got=%d", len(ts))
				}
			},
		},
		{
			"one old call",
			[]time.Time{now.Add(-time.Minute - time.Millisecond)},
			now,
			func(t *testing.T, ts []time.Time) {
				if len(ts) != 0 {
					t.Errorf("unexpected number of timestamps want=0, got=%d delays: %v", len(ts), calculateRelativeTS(now, ts))
				}
			},
		},
		{
			"old calls",
			[]time.Time{now.Add(-time.Hour), now.Add(-time.Minute * 2), now.Add(-time.Minute - time.Millisecond)},
			now,
			func(t *testing.T, ts []time.Time) {
				if len(ts) != 0 {
					t.Errorf("unexpected number of timestamps want=0, got=%d delays: %v", len(ts), calculateRelativeTS(now, ts))
				}
			},
		},
		{
			"old and recent calls",
			[]time.Time{now.Add(-time.Hour), now.Add(-time.Minute + time.Second), now.Add(-time.Second)},
			now,
			func(t *testing.T, ts []time.Time) {
				if len(ts) != 2 {
					t.Errorf("unexpected number of timestamps want=2, got=%d delays: %v", len(ts), calculateRelativeTS(now, ts))
				}
				if now.Sub(ts[0]) > time.Minute {
					t.Errorf("unexpected timestamp %v", now.Sub(ts[0]))
				}
			},
		},
		{
			"recent calls",
			[]time.Time{now.Add(-time.Minute), now.Add(-time.Minute + time.Second), now.Add(-time.Millisecond)},
			now,
			func(t *testing.T, ts []time.Time) {
				if len(ts) != 3 {
					t.Errorf("unexpected number of timestamps want=3, got=%d delays: %v", len(ts), calculateRelativeTS(now, ts))
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := RateLimitedFetcher[any]{
				timestamps: tt.timestamps,
			}
			f.pruneCallsLocked(tt.ts)
			tt.inspect(t, f.timestamps)
		})
	}
}

func Test_nextCallDue(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		timestamps []time.Time
		ts         time.Time
		rateLimit  int
		want       time.Duration
	}{
		{
			"no previous calls",
			nil,
			now,
			3,
			0,
		},
		{
			"recent calls",
			[]time.Time{now.Add(-time.Minute + time.Second), now.Add(-time.Millisecond)},
			now,
			3,
			0,
		},
		{
			"too many calls",
			[]time.Time{now.Add(-time.Second * 4), now.Add(-time.Second * 2), now.Add(-time.Second)},
			now,
			3,
			time.Minute - time.Second*4, // First call was 4 seconds ago. Next call has to wait until we "forget" it.
		},
		{
			"too many calls with a burst",
			[]time.Time{now.Add(-time.Second * 30), now.Add(-time.Second * 3), now.Add(-time.Second * 2), now.Add(-time.Second)},
			now,
			2,
			time.Minute - time.Second*3, // First call was 4 seconds ago. Next call has to wait until we "forget" it.
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := RateLimitedFetcher[any]{
				timestamps: tt.timestamps,
				RateLimit:  tt.rateLimit,
			}
			got := f.nextCallDue(tt.ts)
			if got != tt.want {
				t.Errorf("nextCallDue failed: want=%v got=%v", tt.want, got)
			}
		})
	}
}
