//go:build !solution

package reverse

import (
	"strings"
	"unicode/utf8"
)

func Reverse(input string) string {
	var ans strings.Builder
	ans.Grow(len(input))
	for i := len(input); i > 0; {
		runeValue, width := utf8.DecodeLastRuneInString(input[:i])
		ans.WriteRune(runeValue)
		i -= width
	}
	return ans.String()
}
