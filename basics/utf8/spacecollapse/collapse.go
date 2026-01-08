//go:build !solution

package spacecollapse

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type Space int

const (
	Regular Space = iota
	SkipSpace
)

func CollapseSpaces(input string) string {
	var ans strings.Builder
	ans.Grow(len(input))

	var state = Regular

	for i := 0; i < len(input); {
		runeValue, width := utf8.DecodeRuneInString(input[i:])
		if unicode.IsSpace(runeValue) && state == Regular {
			ans.WriteRune(' ')
			state = SkipSpace
		} else if !unicode.IsSpace(runeValue) {
			if state == SkipSpace {
				state = Regular
			}
			ans.WriteRune(runeValue)
		}

		i += width
	}

	return ans.String()
}
