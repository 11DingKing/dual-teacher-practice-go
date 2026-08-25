package clock

import (
	"testing"
	"time"
)

func TestFixedClock(t *testing.T) {
	want := time.Date(2026, 8, 25, 1, 2, 3, 0, time.UTC)
	if got := (Fixed{T: want}).Now(); !got.Equal(want) {
		t.Fatalf("%v", got)
	}
}
func TestRealClockUTC(t *testing.T) {
	if (Real{}).Now().Location() != time.UTC {
		t.Fatal("clock must be UTC")
	}
}
