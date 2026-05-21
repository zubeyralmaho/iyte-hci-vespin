package partysessions

import "testing"

func TestLegalTransition(t *testing.T) {
	tests := []struct {
		from string
		to   string
		want bool
	}{
		// From active.
		{"active", "active", true},
		{"active", "paused", true},
		{"active", "ended", true},

		// From paused.
		{"paused", "active", true},
		{"paused", "paused", true},
		{"paused", "ended", true},

		// Ended is terminal; only same-status is allowed.
		{"ended", "ended", true},
		{"ended", "active", false},
		{"ended", "paused", false},

		// Unknown statuses are always rejected, on either side.
		{"", "active", false},
		{"active", "", false},
		{"active", "stopped", false},
		{"queued", "active", false},
	}
	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			if got := legalTransition(tt.from, tt.to); got != tt.want {
				t.Errorf("legalTransition(%q, %q) = %v, want %v",
					tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestValidStatus(t *testing.T) {
	for _, s := range []string{"active", "paused", "ended"} {
		if !validStatus(s) {
			t.Errorf("validStatus(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"", "Active", "ACTIVE", "stopped", "queued"} {
		if validStatus(s) {
			t.Errorf("validStatus(%q) = true, want false", s)
		}
	}
}
