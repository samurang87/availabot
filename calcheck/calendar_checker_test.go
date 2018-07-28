package calcheck

import (
	"google.golang.org/api/calendar/v3"
	"testing"
	"time"
)

func TestGetNextThreeEveningsFromAfternoon(t *testing.T) {

	// fixture
	afternoonStart, _ := time.Parse(time.RFC3339, "2018-02-15T15:04:05+01:00")

	firstSlot := calendar.TimePeriod{
		Start: "2018-02-15T16:04:05+01:00",
		End:   "2018-02-15T17:04:05+01:00",
	}

	secondSlot := calendar.TimePeriod{
		Start: "2018-02-16T18:04:05+01:00",
		End:   "2018-02-16T21:04:05+01:00",
	}

	calendarExample := calendar.FreeBusyCalendar{Busy: []*calendar.TimePeriod{&firstSlot, &secondSlot}}

	freeDates := make([]time.Time, 3)
	freeDates[0], _ = time.Parse(time.RFC3339, "2018-02-15T19:00:00+01:00")
	freeDates[1], _ = time.Parse(time.RFC3339, "2018-02-17T19:00:00+01:00")
	freeDates[2], _ = time.Parse(time.RFC3339, "2018-02-18T19:00:00+01:00")

	// execution
	result, err := GetNextThreeEvenings(afternoonStart, calendarExample.Busy)

	if err != nil {
		panic(err)
	}

	// check
	for i, want := range freeDates {

		got := result[i]

		if want.Unix() != got.Unix() {
			t.Errorf("For date n. %v, wanted %v, got %v", i, want, got)
		}
	}

}

func TestGetNextThreeEveningsFromEvening(t *testing.T) {

	// fixture
	afternoonStart, _ := time.Parse(time.RFC3339, "2018-02-15T20:04:05+01:00")

	onlySlot := calendar.TimePeriod{
		Start: "2018-02-16T18:04:05+01:00",
		End:   "2018-02-16T21:04:05+01:00",
	}

	calendarExample := calendar.FreeBusyCalendar{Busy: []*calendar.TimePeriod{&onlySlot}}

	freeDates := make([]time.Time, 3)
	freeDates[0], _ = time.Parse(time.RFC3339, "2018-02-17T19:00:00+01:00")
	freeDates[1], _ = time.Parse(time.RFC3339, "2018-02-18T19:00:00+01:00")
	freeDates[2], _ = time.Parse(time.RFC3339, "2018-02-19T19:00:00+01:00")

	// execution
	result, err := GetNextThreeEvenings(afternoonStart, calendarExample.Busy)

	if err != nil {
		panic(err)
	}

	// check
	for i, want := range freeDates {

		got := result[i]

		if want.Unix() != got.Unix() {
			t.Errorf("For date n. %v, wanted %v, got %v", i, want, got)
		}
	}

}

func TestGetNextThreeEveningsButYouAreOnVacation(t *testing.T) {

	// fixture
	afternoonStart, _ := time.Parse(time.RFC3339, "2018-02-15T20:04:05+01:00")

	onlySlot := calendar.TimePeriod{
		Start: "2018-02-16T18:04:05+01:00",
		End:   "2018-02-28T21:04:05+01:00",
	}

	calendarExample := calendar.FreeBusyCalendar{Busy: []*calendar.TimePeriod{&onlySlot}}

	// execution
	result, err := GetNextThreeEvenings(afternoonStart, calendarExample.Busy)

	if err != nil {
		panic(err)
	}

	// check
	if len(result) != 0 {
		t.Errorf("It should not have returned any free evenings! But it returned %v", result)

	}

}
