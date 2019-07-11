package ical

import (
	"fmt"
	"time"
)

// validateCalendar validate calendar props
func (p *parser) validateCalendar(c *Calendar) error {
	requiredCount := 0
	for _, prop := range c.Properties {
		if prop.Name == "PRODID" {
			c.Prodid = prop.Value
			requiredCount++
		}

		if prop.Name == "VERSION" {
			c.Version = prop.Value
			requiredCount++
		}

		if prop.Name == "CALSCALE" {
			c.Calscale = prop.Value
		}

		if prop.Name == "METHOD" {
			c.Method = prop.Value
		}
	}

	if requiredCount != 2 {
		return fmt.Errorf("missing either required property \"prodid / version /\"")
	}

	return nil
}

// validateEvent validate event props
func (p *parser) validateEvent(v *Event) error {
	uniqueCount := make(map[string]int)

	for _, prop := range v.Properties {
		if prop.Name == "UID" {
			v.UID = prop.Value
			uniqueCount["UID"]++
		}

		if prop.Name == "DTSTAMP" {
			v.Timestamp, _ = parseDate(prop, p.location)
			uniqueCount["DTSTAMP"]++
		}

		if prop.Name == "DTSTART" {
			v.StartDate, _ = parseDate(prop, p.location)
			uniqueCount["DTSTART"]++
		}

		if prop.Name == "DTEND" {
			if hasProperty("DURATION", v.Properties) {
				return fmt.Errorf("Either \"dtend\" or \"duration\" MAY appear")
			}
			v.EndDate, _ = parseDate(prop, p.location)
			uniqueCount["DTEND"]++
		}

		if prop.Name == "DURATION" {
			if hasProperty("DTEND", v.Properties) {
				return fmt.Errorf("Either \"dtend\" or \"duration\" MAY appear")
			}
			uniqueCount["DURATION"]++
		}

		if prop.Name == "SUMMARY" {
			v.Summary = prop.Value
			uniqueCount["SUMMARY"]++
		}

		if prop.Name == "DESCRIPTION" {
			v.Description = prop.Value
			uniqueCount["DESCRIPTION"]++
		}
	}

	if p.c.Method == "" && v.Timestamp.IsZero() {
		return fmt.Errorf("missing required property \"dtstamp\"")
	}

	if v.UID == "" {
		return fmt.Errorf("missing required property \"uid\"")
	}

	if v.StartDate.IsZero() {
		return fmt.Errorf("missing required property \"dtstart\"")
	}

	for key, value := range uniqueCount {
		if value > 1 {
			return fmt.Errorf("\"%s\" property must not occur more than once", key)
		}
	}

	if !hasProperty("DTEND", v.Properties) {
		v.EndDate = v.StartDate.Add(time.Hour * 24) // add one day to start date
	}

	return nil
}

// validateAlarm validate alarm props
func (p *parser) validateAlarm(a *Alarm) error {
	requiredCount := 0
	uniqueCount := make(map[string]int)
	for _, prop := range a.Properties {
		if prop.Name == "ACTION" {
			a.Action = prop.Value
			requiredCount++
			uniqueCount["ACTION"]++
		}

		if prop.Name == "TRIGGER" {
			a.Trigger = prop.Value
			requiredCount++
			uniqueCount["TRIGGER"]++
		}
	}

	if requiredCount != 2 {
		return fmt.Errorf("missing either required property \"action / trigger /\"")
	}

	for key, value := range uniqueCount {
		if value > 1 {
			return fmt.Errorf("\"%s\" property must not occur more than once", key)
		}
	}

	return nil
}

func (p *parser) validateTimezone(a *Timezone) error {
	return nil
}

func (p *parser) validateStandard(a *Standard) error {
	return nil
}

func (p *parser) validateDaylight(a *Daylight) error {
	return nil
}
