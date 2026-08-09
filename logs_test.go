package logs

import (
	"testing"
	"time"
)

func TestAppendDate(t *testing.T) {
	timeValue := time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		date   date
		format dateFormat
		want   string
	}{
		{"day-month-year", DATE_DAY_MONTH_YEAR, DATEFORMAT_DAY_MONTH_YEAR, "20/08/2026"},
		{"month-day-year", DATE_DAY_MONTH_YEAR, DATEFORMAT_MONTH_DAY_YEAR, "08/20/2026"},
		{"year-month-day", DATE_DAY_MONTH_YEAR, DATEFORMAT_YEAR_MONTH_DAY, "2026/08/20"},
		{"day-month", DATE_DAY_MONTH, DATEFORMAT_DAY_MONTH_YEAR, "20/08"},
		{"month-day", DATE_DAY_MONTH, DATEFORMAT_MONTH_DAY_YEAR, "08/20"},
		{"month-year", DATE_MONTH_YEAR, DATEFORMAT_DAY_MONTH_YEAR, "08/2026"},
		{"year-month", DATE_MONTH_YEAR, DATEFORMAT_YEAR_MONTH_DAY, "2026/08"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			previous := logger
			logger.Date = test.date
			logger.DateFormat = test.format
			t.Cleanup(func() { logger = previous })

			var buffer []byte
			appendDate(&buffer, timeValue)

			got := string(buffer)
			want := gray + test.want
			if got != want {
				t.Fatalf("appendDate() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestAppendTimer(t *testing.T) {
	timeValue := time.Date(2026, time.August, 9, 14, 5, 6, 789123000, time.UTC)
	tests := []struct {
		name         string
		timer        timer
		secondFormat secondPrecision
		want         string
	}{
		{"all-seconds", TIMER_HOUR | TIMER_MINUTE | TIMER_SECOND, SECPRECISION_SECOND, "14h 5m 6s"},
		{"hour-minute", TIMER_HOUR | TIMER_MINUTE, SECPRECISION_SECOND, "14h 5m"},
		{"milliseconds", TIMER_SECOND, SECPRECISION_MILLI, "6.789ms"},
		{"microseconds", TIMER_SECOND, SECPRECISION_MICRO, "6.789123us"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			previous := logger
			logger.Timer = test.timer
			logger.SecondFormat = test.secondFormat
			t.Cleanup(func() { logger = previous })

			var buffer []byte
			appendTimer(&buffer, timeValue)

			got := string(buffer)
			want := gray + test.want
			if got != want {
				t.Fatalf("appendTimer() = %q, want %q", got, test.want)
			}
		})
	}
}
