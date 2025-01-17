# iCalendar lexer/parser

[![Build Status](https://travis-ci.org/luxifer/ical.svg?branch=master)](https://travis-ci.org/luxifer/ical)

Golang iCalendar lexer/parser implementing [RFC 5545](https://tools.ietf.org/html/rfc5545). This project is heavily inspired of the talk [Lexical Scanning in Go](https://www.youtube.com/watch?v=HxaD_trXwRE) by Rob Pike.

## Usage

```go
import (
    "github.com/iswangwenbin/ical"
)

// filename is an io.Reader
// second parameter is a *time.Location which defaults to system local
calendar, err := ical.Parse(filename, nil)
```

## Components

| Component | Reference | Status |
|---|---|---|
| VCALENDAR | [RFC5545.Section 3.4](https://tools.ietf.org/html/rfc5545#section-3.4)     |  ✓ 
| VEVENT    | [RFC5545.Section 3.6.1](https://tools.ietf.org/html/rfc5545#section-3.6.1) |  ✓  
| VALARM    | [RFC5545.Section 3.6.6](https://tools.ietf.org/html/rfc5545#section-3.6.6) |  ✓
| VTIMEZONE | [RFC5545.Section 3.6.5](https://tools.ietf.org/html/rfc5545#section-3.6.5) |  ✓
| STANDARD  | [RFC5545.Section 3.6.5](https://tools.ietf.org/html/rfc5545#section-3.6.5) |  ✓
| DAYLIGHT  | [RFC5545.Section 3.6.5](https://tools.ietf.org/html/rfc5545#section-3.6.5) |  ✓ 
| VTODO     | [RFC5545.Section 3.6.2](https://tools.ietf.org/html/rfc5545#section-3.6.2) |
| VJOURNAL  | [RFC5545.Section 3.6.3](https://tools.ietf.org/html/rfc5545#section-3.6.3) |
| VFREEBUSY | [RFC5545.Section 3.6.4](https://tools.ietf.org/html/rfc5545#section-3.6.4) |

## TODO

* [x] Implements VEVENT
* [x] Implements VALARM
* [x] Implements VTIMEZONE
* [x] Implements STANDARD
* [x] Implements DAYLIGHT
* [ ] Implements VTODO
* [ ] Implements VJOURNAL
* [ ] Implements VFREEBUSY
* [ ] Implements Missing Properties on VEVENT
* [ ] Implements Missing Properties on VTIMEZONE
 
