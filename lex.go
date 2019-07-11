package ical

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	// Special tokens
	itemError itemType = iota
	itemEOF
	itemLineEnd

	// Properties
	itemName
	itemParamName
	itemParamValue
	itemValue

	// Punctuation
	itemColon      // :
	itemSemiColon  // ;
	itemEqual      // =
	itemComma      // ,

	// Keyword
	itemKeyword  // delimit the keyword list

	// Delimiters
	itemBeginVCalendar  // BEGIN:VCALENDAR
	itemEndVCalendar    // END:VCALENDAR
	itemBeginVEvent     // BEGIN:VEVENT
	itemEndVEvent       // END:VEVENT
	itemBeginVAlarm     // BEGIN:VALARM
	itemEndVAlarm       // END:VALARM
	itemBeginVTimezone  // BEGIN:VTIMEZONE
	itemEndVTimezone    // END:VTIMEZONE
	itemBeginStandard   // BEGIN:STANDARD
	itemEndStandard     // END:STANDARD
	itemBeginDaylight   // BEGIN:DAYLIGHT
	itemEndDaylight     // END:DAYLIGHT
)

var key = map[string]itemType{
	"BEGIN:VCALENDAR": itemBeginVCalendar,
	"END:VCALENDAR":   itemEndVCalendar,
	"BEGIN:VEVENT":    itemBeginVEvent,
	"END:VEVENT":      itemEndVEvent,
	"BEGIN:VALARM":    itemBeginVAlarm,
	"END:VALARM":      itemEndVAlarm,
	"BEGIN:VTIMEZONE": itemBeginVTimezone,
	"END:VTIMEZONE":   itemEndVTimezone,
	"BEGIN:STANDARD":  itemBeginStandard,
	"END:STANDARD":    itemEndStandard,
	"BEGIN:DAYLIGHT":  itemBeginDaylight,
	"END:DAYLIGHT":    itemEndDaylight,
}

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	input   string    // the string being scanned
	items   chan item // channel of scanned items
	state   stateFn   // the next lexing function to enter
	start   int       // start position of this item
	pos     int       // current position in the input
	width   int       // width of last rune read from input
	lastPos int       // position of most recent item returned by nextItem
}

// lex creates a new scanner for the input string.
func lex(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	go l.run() // Concurrently run state machine.
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexName; l.state != nil; {
		l.state = l.state(l)
	}
	close(l.items) // No more tokens will be delivered.
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	if debug {
		fmt.Printf("string: %+v [%s]\n", item{t, l.start, l.input[l.start:l.pos]}, l.input[l.start:l.pos])
		fmt.Print("emit(): ", " start:", l.start, " pos:", l.pos, " t:", t, "\n\n")
	}
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	if debug {
		fmt.Printf("peek(): %v %d %c\n", r, len(string(r)), r)
	}
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextItem() item {
	item := <-l.items
	if debug {
		fmt.Printf("{{ %+v }}\n", item)
	}
	l.lastPos = item.pos
	return item
}

// State functions

const (
	crlf           = "\r\n"
	beginVCalendar = "BEGIN:VCALENDAR"
	endVCalendar   = "END:VCALENDAR"
	beginVEvent    = "BEGIN:VEVENT"
	endVEvent      = "END:VEVENT"
	beginValarm    = "BEGIN:VALARM"
	endVAlarm      = "END:VALARM"
	beginVTimezone = "BEGIN:VTIMEZONE"
	endVTimezone   = "END:VTIMEZONE"
	beginStandard  = "BEGIN:STANDARD"
	endStandard    = "END:STANDARD"
	beginDaylight  = "BEGIN:DAYLIGHT"
	endDaylight    = "END:DAYLIGHT"
)

func lexContentLine(l *lexer) stateFn {
	switch r := l.next(); {
	case r == ';':
		l.emit(itemSemiColon)
		return lexParamName
	case r == ',':
		l.emit(itemComma)
		return lexParamValue
	case r == ':':
		l.emit(itemColon)
		return lexValue
	default:
		return l.errorf("unrecognized character in action: %#U", r)
	}
}

// lexNewLine scans CRLF
func lexNewLine(l *lexer) stateFn {
	if l.peek() == eof {
		return nil
	}

	if !strings.HasPrefix(l.input[l.pos:], crlf) {
		l.errorf("unable to find end of line \"CRLF\"")
	}

	l.pos += len(crlf)
	l.emit(itemLineEnd)

	if l.next() == eof {
		l.emit(itemEOF)
		return nil
	}
	l.backup()

	return lexName
}

// lexName scans the name in the content line
//
// name       = iana-token / x-name
// iana-token = 1*(ALPHA / DIGIT / "-") ; iCalendar identifier registered with IANA
// x-name     = "X-" [vendorid "-"] 1*(ALPHA / DIGIT / "-") ; Reserved for experimental use.
// vendorid   = 3*(ALPHA / DIGIT) ; Vendor identification
func lexName(l *lexer) stateFn {

	if debug {
		fmt.Println("\n\n\nlexName(): ", " start:", l.start, " pos:", l.pos, " width:", l.width, " lastPos:", l.lastPos, " len:", len(l.input))
	}

	// BEGIN:VCALENDAR
	if strings.HasPrefix(l.input[l.pos:], beginVCalendar) {
		l.pos += len(beginVCalendar)
		l.emit(itemBeginVCalendar)
		if debug {
			fmt.Println("lexNewLine(): ", beginVCalendar)
		}
		return lexNewLine
	}

	// END:VCALENDAR
	if strings.HasPrefix(l.input[l.pos:], endVCalendar) {
		l.pos += len(endVCalendar)
		l.emit(itemEndVCalendar)
		if debug {
			fmt.Println("lexNewLine(): ", endVCalendar)
		}
		return lexNewLine
	}

	// BEGIN:VEVENT
	if strings.HasPrefix(l.input[l.pos:], beginVEvent) {
		l.pos += len(beginVEvent)
		l.emit(itemBeginVEvent)
		if debug {
			fmt.Println("lexNewLine(): ", endVCalendar)
		}
		return lexNewLine
	}

	// END:VEVENT
	if strings.HasPrefix(l.input[l.pos:], endVEvent) {
		l.pos += len(endVEvent)
		l.emit(itemEndVEvent)
		if debug {
			fmt.Println("lexNewLine(): ", endVCalendar)
		}
		return lexNewLine
	}

	// BEGIN:VALARM
	if strings.HasPrefix(l.input[l.pos:], beginValarm) {
		l.pos += len(beginValarm)
		l.emit(itemBeginVAlarm)
		if debug {
			fmt.Println("lexNewLine(): ", endVCalendar)
		}
		return lexNewLine
	}

	// END:VALARM
	if strings.HasPrefix(l.input[l.pos:], endVAlarm) {
		l.pos += len(endVAlarm)
		l.emit(itemEndVAlarm)
		if debug {
			fmt.Println("lexNewLine(): ", endVCalendar)
		}
		return lexNewLine
	}

	// BEGIN:VTIMEZONE
	if strings.HasPrefix(l.input[l.pos:], beginVTimezone) {
		l.pos += len(beginVTimezone)
		l.emit(itemBeginVTimezone)
		if debug {
			fmt.Println("lexNewLine(): ", beginVTimezone)
		}
		return lexNewLine
	}

	// END:VTIMEZONE
	if strings.HasPrefix(l.input[l.pos:], endVTimezone) {
		l.pos += len(endVTimezone)
		l.emit(itemEndVTimezone)
		if debug {
			fmt.Println("lexNewLine(): ", endVTimezone)
		}
		return lexNewLine
	}

	// BEGIN:STANDARD
	if strings.HasPrefix(l.input[l.pos:], beginStandard) {
		l.pos += len(beginStandard)
		l.emit(itemBeginStandard)
		if debug {
			fmt.Println("lexNewLine(): ", beginStandard)
		}
		return lexNewLine
	}

	// END:STANDARD
	if strings.HasPrefix(l.input[l.pos:], endStandard) {
		l.pos += len(endStandard)
		l.emit(itemEndStandard)
		if debug {
			fmt.Println("lexNewLine(): ", endStandard)
		}
		return lexNewLine
	}

	// BEGIN:DAYLIGHT
	if strings.HasPrefix(l.input[l.pos:], beginDaylight) {
		l.pos += len(beginDaylight)
		l.emit(itemBeginDaylight)
		if debug {
			fmt.Println("lexNewLine(): ", beginDaylight)
		}
		return lexNewLine
	}

	// END:DAYLIGHT
	if strings.HasPrefix(l.input[l.pos:], endDaylight) {
		l.pos += len(endDaylight)
		l.emit(itemEndDaylight)
		if debug {
			fmt.Println("lexNewLine(): ", endDaylight)
		}
		return lexNewLine
	}

Loop:
	for {
		switch r := l.next(); {
		case isName(r):
			// absorb
			if debug {
				fmt.Println("isName(): ", "[", r, "]", "<", string(r), ">")
			}
		default:
			if debug {
				fmt.Println("isName(): ", "[", r, "]", "<", string(r), ">")
			}
			l.backup()
			l.emit(itemName)
			break Loop
		}
	}

	if debug {
		fmt.Println("lexContentLine()")
	}

	return lexContentLine
}

// lexParamName scans the param-name in the content line
//
// param-name = iana-token / x-name
// iana-token = 1*(ALPHA / DIGIT / "-") ; iCalendar identifier registered with IANA
// x-name     = "X-" [vendorid "-"] 1*(ALPHA / DIGIT / "-") ; Reserved for experimental use.
// vendorid   = 3*(ALPHA / DIGIT) ; Vendor identification
func lexParamName(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isName(r):
			// absorb
		default:
			l.backup()
			l.emit(itemParamName)
			break Loop
		}
	}
	r := l.next()
	if r == '=' {
		l.emit(itemEqual)
		return lexParamValue
	}
	return l.errorf("missing \"=\" sign after param name, got %#U", r)
}

// lexParamValue scans the param-value in the content line
//
// param-value   = paramtext / quoted-string
// paramtext     = *SAFE-CHAR
// quoted-string = DQUOTE *QSAFE-CHAR DQUOTE
// QSAFE-CHAR    = WSP / %x21 / %x23-7E / NON-US-ASCII ; Any character except CONTROL and DQUOTE
// SAFE-CHAR     = WSP / %x21 / %x23-2B / %x2D-39 / %x3C-7E / NON-US-ASCII ; Any character except CONTROL, DQUOTE, ";", ":", ","
func lexParamValue(l *lexer) stateFn {
	r := l.next()

	if r == '"' {
		l.ignore()
	QLoop:
		for {
			switch r := l.next(); {
			case isQSafeChar(r):
				// absorb
			default:
				l.backup()
				l.emit(itemParamValue)
				break QLoop
			}
		}
		r := l.next()
		if r != '"' {
			l.errorf("Missing \" for closing value")
		} else {
			l.ignore()
		}
	} else {
		l.backup()
	Loop:
		for {
			switch r := l.next(); {
			case isSafeChar(r):
				// absorb
			default:
				l.backup()
				l.emit(itemParamValue)
				break Loop
			}
		}
	}
	return lexContentLine
}

// lexValue scans the value in the content line
//
// value      = *VALUE-CHAR
// VALUE-CHAR = WSP / %x21-7E / NON-US-ASCII ; Any textual character
func lexValue(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isValueChar(r):
			// absorb
		default:
			l.backup()
			l.emit(itemValue)
			break Loop
		}
	}
	return lexNewLine
}
