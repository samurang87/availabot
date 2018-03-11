package calendar_checker

import (
	"testing"
	"time"
	"google.golang.org/api/calendar/v3"
)

func TestGetNextThreeEvenings(t *testing.T) {

	// fixture
	afternoon_start, _ := time.Parse(time.RFC3339, "2018-02-15T15:04:05+01:00")

	first_slot := calendar.TimePeriod{
		Start: "2018-02-15T16:04:05+01:00",
		End: "2018-02-15T17:04:05+01:00",
	}

	second_slot := calendar.TimePeriod{
		Start: "2018-02-16T18:04:05+01:00",
		End: "2018-02-16T21:04:05+01:00",
	}

	calendar_example := calendar.FreeBusyCalendar{Busy:[]*calendar.TimePeriod{&first_slot, &second_slot}}

	free_dates := make([]time.Time, 3)
	free_dates[0], _ = time.Parse(time.RFC3339, "2018-02-15T19:00:00+01:00")
	free_dates[1], _ = time.Parse(time.RFC3339, "2018-02-17T19:00:00+01:00")
	free_dates[2], _ = time.Parse(time.RFC3339, "2018-02-18T19:00:00+01:00")

	// execution
	result := GetNextThreeEvenings(afternoon_start, calendar_example.Busy)


	// check
	for i, want := range(free_dates) {
		
		got := result[i]

		if want != got {
			t.Errorf("For date n. %v, wanted %v, got %v", i, want, got)
		}
	}

}
