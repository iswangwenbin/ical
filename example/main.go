package main

import (
	"encoding/json"
	"fmt"
	"github.com/iswangwenbin/ical"
	"github.com/tidwall/pretty"
	"io/ioutil"
	"strings"
)

func main() {
	filename := "../fixtures/icalendar.ics"
	//filename = "../fixtures/work.ics"
	fmt.Println(filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}
	calendar, err := ical.Parse(strings.NewReader(string(data)), nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("{{ Property }}")

	for _, prop := range calendar.Properties {
		data, err := json.Marshal(prop)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("%s\n", pretty.Pretty(data))
	}

	fmt.Println("{{ VEVENT }}")

	for _, event := range calendar.Events {
		data, err := json.Marshal(event)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("%s\n", pretty.Pretty(data))
	}

	fmt.Println("{{ VTIMEZONE }}")

	for _, tz := range calendar.Timezones {
		data, err := json.Marshal(tz)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("%s\n", pretty.Pretty(data))
	}

}
