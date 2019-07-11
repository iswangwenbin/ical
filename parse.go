package ical

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"
)

type parser struct {
	lex       *lexer
	token     [2]item
	peekCount int
	scope     int
	c         *Calendar
	v         *Event
	a         *Alarm
	t         *Timezone
	s         *Standard
	d         *Daylight
	location  *time.Location
}

// Parse transforms the raw iCalendar into a Calendar struct
// It's up to the caller to close the io.Reader
// if the time.Location parameter is not set, it will default to the system location
func Parse(r io.Reader, l *time.Location) (*Calendar, error) {
	p := &parser{}
	p.c = NewCalendar()
	p.scope = scopeCalendar
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// timezone
	if l == nil {
		l = time.Local
	}
	p.location = l

	// 	single line
	text := unfold(string(bytes))
	p.lex = lex(text)

	return p.parse()
}

// unfold convert multiple line value to one line
// from rfc5545-3.1
// a long line can be split between any two characters by inserting a CRLF
// immediately followed by a single linear white-space character (i.e., SPACE or HTAB).
func unfold(text string) string {
	return strings.NewReplacer("\r\n ", "", "\r\n\t", "").Replace(text)
}

// next returns the next token.
func (p *parser) next() item {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.token[0] = p.lex.nextItem()
	}
	return p.token[p.peekCount]
}

// backup backs the input stream up one token.
func (p *parser) backup() {
	p.peekCount++
}

// peek returns but does not consume the next token.
func (p *parser) peek() item {
	if p.peekCount > 0 {
		return p.token[p.peekCount-1]
	}
	p.peekCount = 1
	p.token[0] = p.lex.nextItem()
	return p.token[0]
}

// enterScope switch scope between Calendar, Event and Alarm
func (p *parser) enterScope(scope int) {
	p.scope = scope
	// p.scope++
}

// leaveScope returns to previous scope
func (p *parser) leaveScope(scope int) {
	if scope == -1 {
		p.scope--
	} else {
		p.scope = scope
	}
}

// parse

const (
	scopeCalendar int = iota
	scopeEvent
	scopeAlarm
	scopeTimezone
	scopeStandard
	scopeDaylight
)

const (
	dateLayout              = "20060102"
	dateTimeLayoutUTC       = "20060102T150405Z"
	dateTimeLayoutLocalized = "20060102T150405"
)

var errorDone = errors.New("done")

func (p *parser) parse() (*Calendar, error) {
	if item := p.next(); item.typ != itemBeginVCalendar {
		return nil, fmt.Errorf("found %s, expected BEGIN:VCALENDAR", item)
	}

	if item := p.next(); item.typ != itemLineEnd {
		return nil, fmt.Errorf("found %s, expected CRLF", item)
	}

	for {
		err := p.scanContentLine()
		if err == errorDone {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return p.c, nil
}

// scanDelimiter switch scope and validate related component
func (p *parser) scanDelimiter(delim item) error {
	if delim.typ == itemBeginVEvent {
		if err := p.validateCalendar(p.c); err != nil {
			return err
		}

		p.v = NewEvent()
		p.enterScope(scopeEvent)

		if item := p.next(); item.typ != itemLineEnd {
			return fmt.Errorf("found %s, expected CRLF", item)
		}
	}

	if delim.typ == itemEndVEvent {
		//if p.scope > scopeEvent {
		//	return fmt.Errorf("found %s, expeced END:VALARM", delim)
		//}

		if err := p.validateEvent(p.v); err != nil {
			return err
		}

		p.c.Events = append(p.c.Events, p.v)
		p.leaveScope(scopeCalendar)

		if item := p.next(); item.typ != itemLineEnd {
			return fmt.Errorf("found %s, expected CRLF", item)
		}
	}

	if delim.typ == itemBeginVTimezone {
		if err := p.validateTimezone(p.t); err != nil {
			return err
		}

		p.t = NewTimezone()
		p.enterScope(scopeTimezone)

		if item := p.next(); item.typ != itemLineEnd {
			return fmt.Errorf("found %s, expected CRLF", item)
		}
	}

	if delim.typ == itemEndVTimezone {
		//if p.scope > scopeEvent {
		//	return fmt.Errorf("found %s, expeced END:VALARM", delim)
		//}

		if err := p.validateTimezone(p.t); err != nil {
			return err
		}

		p.c.Timezones = append(p.c.Timezones, p.t)
		p.leaveScope(scopeCalendar)

		if item := p.next(); item.typ != itemLineEnd {
			return fmt.Errorf("found %s, expected CRLF", item)
		}
	}

	if delim.typ == itemBeginStandard {
		if err := p.validateStandard(p.s); err != nil {
			return err
		}
		p.s = NewStandard()
		p.enterScope(scopeStandard)
		if item := p.next(); item.typ != itemLineEnd {
			return fmt.Errorf("found %s, expected CRLF", item)
		}
	}

	if delim.typ == itemEndStandard {
		if err := p.validateStandard(p.s); err != nil {
			return err
		}
		p.t.Standards = append(p.t.Standards, p.s)
		p.leaveScope(scopeTimezone)
		if item := p.next(); item.typ != itemLineEnd {
			return fmt.Errorf("found %s, expected CRLF", item)
		}
	}

	if delim.typ == itemBeginDaylight {
		if err := p.validateDaylight(p.d); err != nil {
			return err
		}
		p.d = NewDaylight()
		p.enterScope(scopeDaylight)
		if item := p.next(); item.typ != itemLineEnd {
			return fmt.Errorf("found %s, expected CRLF", item)
		}
	}

	if delim.typ == itemEndDaylight {
		if err := p.validateDaylight(p.d); err != nil {
			return err
		}
		p.t.Daylights = append(p.t.Daylights, p.d)
		p.leaveScope(scopeTimezone)
		if item := p.next(); item.typ != itemLineEnd {
			return fmt.Errorf("found %s, expected CRLF", item)
		}
	}

	if delim.typ == itemBeginVAlarm {
		p.a = NewAlarm()
		p.enterScope(scopeAlarm)

		if item := p.next(); item.typ != itemLineEnd {
			return fmt.Errorf("found %s, expected CRLF", item)
		}
	}

	if delim.typ == itemEndVAlarm {
		if err := p.validateAlarm(p.a); err != nil {
			return err
		}

		p.v.Alarms = append(p.v.Alarms, p.a)
		p.leaveScope(scopeEvent)

		if item := p.next(); item.typ != itemLineEnd {
			return fmt.Errorf("found %s, expected CRLF", item)
		}
	}

	if delim.typ == itemEndVCalendar {
		if p.scope > scopeCalendar {
			return fmt.Errorf("found %s, expeced END:VEVENT", delim)
		}
		return errorDone
	}

	return nil
}

// scanContentLine parses a content-line of a calendar
func (p *parser) scanContentLine() error {
	name := p.next()

	if name.typ > itemKeyword {
		if err := p.scanDelimiter(name); err != nil {
			return err
		}
		return p.scanContentLine()
	}

	if !isItemName(name) {
		return fmt.Errorf("found %s, expected a \"name\" token", name)
	}

	prop := NewProperty()
	prop.Name = name.val

	if err := p.scanParams(prop); err != nil {
		return err
	}

	if item := p.next(); item.typ != itemColon {
		return fmt.Errorf("found %s, expected \":\"", item)
	}

	value := p.next()

	if value.typ != itemValue {
		return fmt.Errorf("found %s, expected a value", value)
	}

	prop.Value = value.val

	if item := p.next(); item.typ != itemLineEnd {
		return fmt.Errorf("found %s, expected CRLF", name)
	}

	switch p.scope {
	case scopeCalendar:
		p.c.Properties = append(p.c.Properties, prop)
	case scopeEvent:
		p.v.Properties = append(p.v.Properties, prop)
	case scopeAlarm:
		p.a.Properties = append(p.a.Properties, prop)
	case scopeTimezone:
		p.t.Properties = append(p.t.Properties, prop)
	case scopeDaylight:
		p.d.Properties = append(p.d.Properties, prop)
	case scopeStandard:
		p.s.Properties = append(p.s.Properties, prop)
	default:
		return fmt.Errorf("scope %d, expected =", p.scope)
	}
	return nil
}

// scanParams parses a list of param inside a content-line
func (p *parser) scanParams(prop *Property) error {
	for {
		item := p.next()

		if item.typ != itemSemiColon {
			p.backup()
			return nil
		}

		paramName := p.next()

		if paramName.typ != itemParamName {
			return fmt.Errorf("found %s, expected a param-name", paramName)
		}

		param := NewParam()

		if item := p.next(); item.typ != itemEqual {
			return fmt.Errorf("found %s, expected =", item)
		}

		if err := p.scanValues(param); err != nil {
			return err
		}

		prop.Params[paramName.val] = param
	}
}

// scanValues parses a list of at least one value for a param
func (p *parser) scanValues(param *Param) error {
	paramValue := p.next()

	if paramValue.typ != itemParamValue {
		return fmt.Errorf("found %s, expected a param-value", paramValue)
	}

	param.Values = append(param.Values, paramValue.val)

	for {
		item := p.next()

		if item.typ != itemComma {
			p.backup()
			return nil
		}

		paramValue := p.next()

		if paramValue.typ != itemParamValue {
			return fmt.Errorf("found %s, expected a param-value", paramValue)
		}

		param.Values = append(param.Values, paramValue.val)
	}
}

// hasProperty checks if a given component has a certain property
func hasProperty(name string, properties []*Property) bool {
	for _, prop := range properties {
		if name == prop.Name {
			return true
		}
	}
	return false
}

// parseDate transform an ical date property into a time.Time
func parseDate(prop *Property, l *time.Location) (time.Time, error) {
	if strings.HasSuffix(prop.Value, "Z") {
		return time.Parse(dateTimeLayoutUTC, prop.Value)
	}

	if tz, ok := prop.Params["TZID"]; ok {
		loc, err := time.LoadLocation(tz.Values[0])

		// In case we are not able to load TZID location we default to UTC
		if err != nil {
			loc = time.UTC
		}

		return time.ParseInLocation(dateTimeLayoutLocalized, prop.Value, loc)
	}

	if len(prop.Value) == 8 {
		return time.ParseInLocation(dateLayout, prop.Value, l)
	}

	layout := dateTimeLayoutLocalized

	if val, ok := prop.Params["VALUE"]; ok {
		switch val.Values[0] {
		case "DATE":
			layout = dateLayout
		case "DATE-TIME":
			layout = dateTimeLayoutLocalized
		}
	}

	return time.ParseInLocation(layout, prop.Value, l)
}
