package cmd

import "testing"

func TestDisplayHeaderValueRedactsCredentials(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "Authorization", value: "Bearer private", want: "[REDACTED]"},
		{name: "Cookie", value: "session=private", want: "[REDACTED]"},
		{name: "X-API-Key", value: "private", want: "[REDACTED]"},
		{name: "X-Access-Token", value: "private", want: "[REDACTED]"},
		{name: "Content-Type", value: "application/json", want: "application/json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := displayHeaderValue(tt.name, tt.value); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
