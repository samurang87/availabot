package main

import (
	"github.com/samurang87/availabot/calendar_checker"
	"time"
	"fmt"
)

func main() {

	_, busy, err := calendar_checker.GetBusyCalendar(time.Now())

	if err != nil {
		panic(err)
	}

	threeFree, err := calendar_checker.GetNextThreeEvenings(time.Now(), busy)

	if err != nil {
		panic(err)
	}

	fmt.Println(threeFree)

}
