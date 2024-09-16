package utils

import "strings"

// IndentMultilineString indents a multiline string with the specified number of spaces
func IndentMultilineString(s string, indent int) string {
	indentation := strings.Repeat(" ", indent)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = indentation + line
	}
	return strings.Join(lines, "\n")
}
