package main

import (
	"fmt"
	"github.com/samurang87/availabot/calendar_checker"
	"time"
)

func main() {

	_, busy := calendar_checker.GetBusyCalendar(time.Now())

	for _, slot := range busy {

		fmt.Println(slot.Start)
		fmt.Println(slot.End)

	}
}
