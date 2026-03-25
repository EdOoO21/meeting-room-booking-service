package clock

import "testing"

func TestClock_NowUTC_ReturnsUTC(t *testing.T) {
	t.Parallel()

	got := New().NowUTC()
	if got.Location().String() != "UTC" {
		t.Fatalf("location = %q, want UTC", got.Location())
	}
}
