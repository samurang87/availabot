package main

import (
	"github.com/samurang87/availabot/calendar_checker"
	"time"
	"fmt"
)

func main() {

	_, busy := calendar_checker.GetBusyCalendar(time.Now())

	threeFree := calendar_checker.GetNextThreeEvenings(time.Now(), busy)

	fmt.Println(threeFree)

}
