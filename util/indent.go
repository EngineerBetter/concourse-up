package util

import (
	"strconv"
	"strings"
)

// Indent is a helper function to indent the field a given number of spaces
func Indent(countStr, field string) string {
	count, err := strconv.Atoi(countStr)
	if err != nil {
		panic(err)
	}

	prefix := ""
	for i := 0; i < count; i++ {
		prefix += " "
	}

	field = strings.TrimSpace(field)
	return strings.Replace(field, "\n", "\n"+prefix, -1)
}
