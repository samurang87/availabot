package calendar_checker

import (
	"fmt"
	"testing"
	"time"
)

func TestGetNextThreeEvenings(t *testing.T) {

	afternoon_start, err := time.Parse(time.RFC3339, "2018-02-15T15:04:05+01:00")

	if err == nil {
		fmt.Println(GetNextThreeEvenings(afternoon_start))
	} else {
		fmt.Println(err)
	}



}
