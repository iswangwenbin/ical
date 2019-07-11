package ical

import (
	"unicode"
	"unicode/utf8"
)

// rune helpers

func isName(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-'
}

func isQSafeChar(r rune) bool {
	return !unicode.IsControl(r) && r != '"'
}

func isSafeChar(r rune) bool {
	return !unicode.IsControl(r) && r != '"' && r != ';' && r != ':' && r != ','
}

func isValueChar(r rune) bool {
	return r == '\t' || (!unicode.IsControl(r) && utf8.ValidRune(r))
}

// item helpers

// isItemName checks if the item is an ical name
func isItemName(i item) bool {
	return i.typ == itemName
}
