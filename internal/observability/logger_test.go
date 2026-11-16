package observability

import (
	"errors"
	"testing"
)

func TestSetupLogger(t *testing.T) {
	tests := []struct {
		name   string
		level  string
		format string
	}{
		{"text info", "info", "text"},
		{"json debug", "debug", "json"},
		{"warn", "warn", "text"},
		{"error", "error", "json"},
		{"default fallback", "unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := SetupLogger(tt.level, tt.format)
			if logger == nil {
				t.Fatal("expected non-nil logger")
			}
		})
	}
}

func TestErrorAttr(t *testing.T) {
	err := errors.New("something went wrong")
	attr := ErrorAttr(err)
	if attr.Key != "error" || attr.Value.String() != "something went wrong" {
		t.Errorf("unexpected attr: %v", attr)
	}

	nilAttr := ErrorAttr(nil)
	if nilAttr.Key != "error" || nilAttr.Value.String() != "<nil>" {
		t.Errorf("unexpected nil attr: %v", nilAttr)
	}
}
