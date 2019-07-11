package ical

import (
	"time"
)

// A Calendar represents the whole iCalendar
type Calendar struct {
	Properties []*Property // Properties
	Events     []*Event    // Events
	Timezones  []*Timezone // Timezones
	Prodid     string      // Production Id
	Version    string      // iCalendar version
	Calscale   string      // Calscale: "GREGORIAN"
	Method     string      // Method
}

// An Event represent a VEVENT component in an iCalendar
type Event struct {
	Properties  []*Property
	Alarms      []*Alarm
	UID         string
	Timestamp   time.Time
	StartDate   time.Time
	EndDate     time.Time
	Summary     string
	Description string
}

// An Timezone represent a VTimezone component in an iCalendar
type Timezone struct {
	Properties []*Property
	Standards  []*Standard
	Daylights  []*Daylight
}

// An Standard represent a Standard component in an iCalendar
type Standard struct {
	Properties []*Property
}

// An Daylight represent a Daylight component in an iCalendar
type Daylight struct {
	Properties []*Property
}

// An Alarm represent a VALARM component in an iCalendar
type Alarm struct {
	Properties []*Property
	Action     string
	Trigger    string
}

// A Property represent an unparsed property in an iCalendar component
type Property struct {
	Name   string
	Params map[string]*Param
	Value  string
}

// A Param represent a list of param for a property
type Param struct {
	Values []string
}

// NewCalendar creates an empty Calendar
func NewCalendar() *Calendar {
	c := &Calendar{
		Calscale: "GREGORIAN", // The default value is "GREGORIAN"
	}
	c.Properties = make([]*Property, 0)
	c.Events = make([]*Event, 0)
	c.Timezones = make([]*Timezone, 0)
	return c
}

// NewProperty creates an empty Property
func NewProperty() *Property {
	p := &Property{}
	p.Params = make(map[string]*Param)
	return p
}

// NewParam creates an empty Param
func NewParam() *Param {
	p := &Param{}
	p.Values = make([]string, 0)
	return p
}

// NewEvent creates an empty Event
func NewEvent() *Event {
	v := &Event{}
	v.Properties = make([]*Property, 0)
	v.Alarms = make([]*Alarm, 0)
	return v
}

// NewAlarm creates an empty Alarm
func NewAlarm() *Alarm {
	a := &Alarm{}
	a.Properties = make([]*Property, 0)
	return a
}

// NewTimezone creates an empty Timezone
func NewTimezone() *Timezone {
	v := &Timezone{}
	v.Properties = make([]*Property, 0)
	return v
}

// NewDaylight creates an empty Daylight
func NewDaylight() *Daylight {
	v := &Daylight{}
	v.Properties = make([]*Property, 0)
	return v
}

// NewStandard creates an empty Standard
func NewStandard() *Standard {
	v := &Standard{}
	v.Properties = make([]*Property, 0)
	return v
}
