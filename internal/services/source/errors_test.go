package source

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsRetryable(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"network", &NetworkError{Op: "get", URL: "u", Err: errors.New("reset")}, true},
		{"network wrapped", fmt.Errorf("outer: %w", &NetworkError{Op: "read"}), true},
		{"invalid duration_low", &InvalidAudioError{Reason: ReasonDurationMismatchLow}, true},
		{"invalid too_short", &InvalidAudioError{Reason: ReasonTooShort}, true},
		{"invalid probe_failed", &InvalidAudioError{Reason: ReasonProbeFailed}, true},
		{"invalid duration_high", &InvalidAudioError{Reason: ReasonDurationMismatchHigh}, false},
		{"invalid bitrate_low", &InvalidAudioError{Reason: ReasonBitrateTooLow}, false},
		{"plugin invocation", &PluginInvocationError{PluginEntryPath: "bili"}, false},
		{"all sources failed", &AllSourcesFailedError{Tried: 3}, false},
		{"plain error", errors.New("boom"), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsRetryable(c.err); got != c.want {
				t.Fatalf("IsRetryable(%v) = %v, want %v", c.err, got, c.want)
			}
		})
	}
}
