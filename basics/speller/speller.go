//go:build !solution

package speller

import (
	"strings"
)

func GetGroupRepresentation(course int) string {
	ones := []string{"", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}
	teens := []string{"ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen"}
	tens := []string{"", "", "twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety"}

	if course == 0 {
		return ""
	}

	var result []string

	if course >= 100 {
		result = append(result, ones[course/100], " ")
		result = append(result, "hundred", " ")
		course = course % 100
	}

	if course >= 20 {
		result = append(result, tens[course/10])
		course = course % 10
		if course != 0 {
			result = append(result, "-")
		}
	}

	if course >= 10 {
		result = append(result, teens[course-10])
	} else if course > 0 {
		result = append(result, ones[course])
	}

	return strings.Join(result, "")
}

func Spell(n int64) string {
	if n == 0 {
		return "zero"
	}

	signed := n < 0
	if signed {
		n *= -1
	}

	matchGroup := map[int]string{
		0: "",
		1: "thousand",
		2: "million",
		3: "billion",
		4: "trillion",
		5: "quadrillion",
	}

	var parts []string
	groupIndex := 0

	for n != 0 {
		group := int(n % 1000)
		if group > 0 {
			groupRepresentation := GetGroupRepresentation(group)
			if matchGroup[groupIndex] != "" {
				groupRepresentation += " " + matchGroup[groupIndex]
			}
			parts = append([]string{groupRepresentation}, parts...)
		}
		n /= 1000
		groupIndex++
	}

	if signed {
		parts = append([]string{"minus"}, parts...)
	}

	return strings.TrimSpace(strings.Join(parts, " "))
}
