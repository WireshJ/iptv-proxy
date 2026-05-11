package utils

import "fmt"

// ConvertResponseToString converts an interface response to a string.
func ConvertResponseToString(resp interface{}) string {
	switch v := resp.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
