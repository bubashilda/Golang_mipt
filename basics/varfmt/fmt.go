//go:build !solution

package varfmt

import (
	"fmt"
	"strconv"
	"strings"
)

type State int

const (
	RegularRune State = iota
	CurlyBrace
)

func Sprintf(format string, arguments ...interface{}) string {
	var result strings.Builder
	result.Grow(len(format))

	stringArguments := make([]string, len(arguments))
	for i, argument := range arguments {
		stringArguments[i] = fmt.Sprint(argument)
	}

	parsingState := RegularRune
	consequentIndexFormat := 0

	numStartIndex := 0
	for i, ch := range format {
		if ch == '{' {
			parsingState = CurlyBrace
			numStartIndex = i
		}
		if parsingState == RegularRune && ch != '}' {
			result.WriteRune(ch)
		}
		if ch == '}' {
			parsingState = RegularRune
			num := format[numStartIndex+1 : i]
			val, err := strconv.Atoi(num)
			if err != nil {
				result.WriteString(stringArguments[consequentIndexFormat])
			} else {
				result.WriteString(stringArguments[val])
			}
			consequentIndexFormat += 1
		}
	}

	return result.String()
}
