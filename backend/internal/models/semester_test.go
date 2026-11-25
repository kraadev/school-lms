package models

import (
	"testing"
	"time"
)

func TestSemesterSetStatusUsesInclusiveDateRange(t *testing.T) {
	semester := Semester{
		StartDate: time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 12, 18, 0, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name   string
		now    time.Time
		status string
	}{
		{"before start", time.Date(2026, 7, 12, 23, 59, 59, 0, time.UTC), SemesterStatusComing},
		{"at start", time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC), SemesterStatusOngoing},
		{"during semester", time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC), SemesterStatusOngoing},
		{"at end", time.Date(2026, 12, 18, 23, 59, 59, 0, time.UTC), SemesterStatusOngoing},
		{"after end", time.Date(2026, 12, 19, 0, 0, 0, 0, time.UTC), SemesterStatusEnded},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			semester.SetStatus(test.now)
			if semester.Status != test.status {
				t.Fatalf("status = %q, want %q", semester.Status, test.status)
			}
		})
	}
}
