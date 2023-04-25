// Header parsing functionality.

package eml

import (
	"strings"
)

func split(ts []token, s token) [][]token {
	r, l := [][]token{}, 0
	for i, t := range ts {
		if string(t) == string(s) {
			r = append(r, ts[l:i])
			l = i + 1
		}
	}
	if l != len(ts) {
		r = append(r, ts[l:len(ts)])
	}
	return r
}

// BUG: We don't currently support domain literals with commas.
func parseAddressList(s []byte) ([]Address) {
	al := []Address{}
	ts := tokenize(s)

	tts := split(ts, []byte{','})
	for _, ts := range tts {
		a := parseAddress(ts)
		al = append(al, a)
	}
	return al
}

func decodeRFC2047(word string) string {
	if strings.HasPrefix(word, "=?") && strings.HasSuffix(word, "?=") && strings.Count(word, "?") == 4 {
		return ""
	}
	return word
}
