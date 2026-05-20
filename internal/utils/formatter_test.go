package utils

import (
	"strings"
	"testing"
	"time"
)

func TestFormatPaise(t *testing.T) {
	tests := []struct {
		paise    int64
		expected string
	}{
		{0, "₹0.00"},
		{100, "₹1.00"},
		{150, "₹1.50"},
		{1350000, "₹13500.00"},
		{99, "₹0.99"},
		{1, "₹0.01"},
		{50075, "₹500.75"},
		{-100, "₹-1.00"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := FormatPaise(tt.paise)
			if got != tt.expected {
				t.Errorf("FormatPaise(%d) = %q, want %q", tt.paise, got, tt.expected)
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2026-01-15", "15 Jan 2026"},
		{"2026-12-25", "25 Dec 2026"},
		{"2026-05-20", "20 May 2026"},
		{"invalid-date", "invalid-date"}, // returns input on parse error
		{"", ""},                         // empty returns empty
		{"2026-02-28", "28 Feb 2026"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := FormatDate(tt.input)
			if got != tt.expected {
				t.Errorf("FormatDate(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTodayDate(t *testing.T) {
	got := TodayDate()
	expected := time.Now().Format("2006-01-02")
	if got != expected {
		t.Errorf("TodayDate() = %q, want %q", got, expected)
	}
	// Verify format: YYYY-MM-DD
	if len(got) != 10 || got[4] != '-' || got[7] != '-' {
		t.Errorf("TodayDate() format invalid: %q", got)
	}
}

func TestCurrentTime(t *testing.T) {
	got := CurrentTime()
	// Verify format: HH:MM
	if len(got) != 5 || got[2] != ':' {
		t.Errorf("CurrentTime() format invalid: %q", got)
	}
	// Verify it parses
	_, err := time.Parse("15:04", got)
	if err != nil {
		t.Errorf("CurrentTime() not parseable: %v", err)
	}
}

func TestCurrentMonth(t *testing.T) {
	got := CurrentMonth()
	expected := time.Now().Format("0601")
	if got != expected {
		t.Errorf("CurrentMonth() = %q, want %q", got, expected)
	}
	// Should be 4 digits
	if len(got) != 4 {
		t.Errorf("CurrentMonth() length should be 4, got %d", len(got))
	}
	// All digits
	for _, c := range got {
		if c < '0' || c > '9' {
			t.Errorf("CurrentMonth() contains non-digit: %q", got)
			break
		}
	}
	_ = strings.Contains(got, "") // use strings import
}
