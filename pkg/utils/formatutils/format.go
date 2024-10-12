package formatutils

import (
	"encoding/json"
	"strings"
)

// IndentMultilineString indents a multiline string with the specified number of spaces
func IndentMultilineString(s string, indent int) string {
	indentation := strings.Repeat(" ", indent)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = indentation + line
	}
	return strings.Join(lines, "\n")
}

func RetrieveNestedId(raw_response []byte, group_key string, header_to_match string, value_to_match string, retrieve_field string) (string, error) {
	// Parse the JSON response
	var response map[string]interface{}
	err := json.Unmarshal(raw_response, &response)
	if err != nil {
		return "", err
	}

	var retrieved_field string
	// Iterate through items to find the matching header
	items := response[group_key].([]interface{})
	for _, item := range items {
		if item.(map[string]interface{})[header_to_match] == value_to_match {
			retrieved_field = item.(map[string]interface{})[retrieve_field].(string)
			break
		}
	}
	return retrieved_field, nil
}
