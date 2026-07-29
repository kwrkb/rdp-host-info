package msg

import (
	"testing"

	"github.com/kwrkb/rdp-host-info/internal/msgid"
)

// TestCatalogComplete guards against silent English leaking into Japanese
// output: every ID in msgid.All must resolve to a non-empty string in every
// supported language. A missing catalog entry falls back to the raw ID
// (see Format), which this test also rejects.
func TestCatalogComplete(t *testing.T) {
	for _, id := range msgid.All {
		e, ok := catalog[id]
		if !ok {
			t.Errorf("id %q has no catalog entry", id)
			continue
		}
		if e.en == "" {
			t.Errorf("id %q has empty en entry", id)
		}
		if e.ja == "" {
			t.Errorf("id %q has empty ja entry", id)
		}
	}
}

func TestFormat_UnknownIDFallsBackToRawID(t *testing.T) {
	got := Format(English, msgid.ID("nonexistent"))
	if got != "nonexistent" {
		t.Errorf("Format = %q, want %q", got, "nonexistent")
	}
}

func TestDuration(t *testing.T) {
	tests := []struct {
		lang    Lang
		seconds uint32
		want    string
	}{
		{English, 900, "15 minutes"},
		{English, 60, "1 minute"},
		{English, 90, "90 seconds"},
		{Japanese, 900, "15分"},
		{Japanese, 60, "1分"},
		{Japanese, 90, "90秒"},
	}
	for _, tt := range tests {
		if got := duration(tt.lang, tt.seconds); got != tt.want {
			t.Errorf("duration(%v, %d) = %q, want %q", tt.lang, tt.seconds, got, tt.want)
		}
	}
}
