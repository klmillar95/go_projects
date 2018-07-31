package querytool

import (
	"time"
	"strings"
	"strconv"
)

// gets the integer value for the current time for each index in the cron object to compare to the given value in the cron string
func getCronVal(index int) int {
	val := 0
	layOut := "2006-01-02 15:04:05"

	currentTime := time.Now()
	timeStampString := currentTime.Format("2006-01-02 15:04:05")

	timeStamp, err := time.Parse(layOut, timeStampString)
	hr, min, _ := timeStamp.Clock()

	if err != nil {
		panic(err)
	}

	switch index {
		case 0:
			val = min
		case 1:
			val = hr
		case 2:
			val = currentTime.Day()
		case 3:
			val = int(currentTime.Month())
		case 4:
			val = int(currentTime.Weekday())
	}

	return val
}

// takes a cron string and checks returns true if the current time matches up, or false if it does not
func CheckTime(cron string) bool {
	tokens := strings.Split(cron," ")
	var cron_components [][]string

	for _,token := range tokens {
		cron_components = append(cron_components,strings.Split(token,","))
	}

	for i := 0; i < 5; i++ {
		if cron_components[i][0] != "*" {
			truecount := 0

			for _,comp := range cron_components[i]{
				if len(strings.Split(comp,"-")) > 1 {
					a, _ := strconv.Atoi(strings.Split(comp,"-")[0])
					b, _ := strconv.Atoi(strings.Split(comp,"-")[1])

					if getCronVal(i) >= a && getCronVal(i) <= b {
						truecount++
					}
				} else {
					a, _ := strconv.Atoi(comp)
					if a == getCronVal(i) {
						truecount++
					}
				}
			}

			if truecount == 0 {
				return false
			}
		}
	}

	return true
}
