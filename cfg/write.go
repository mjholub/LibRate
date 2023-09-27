package cfg

import (
	"bytes"
	"fmt"
)

// createKVPairs creates a string of key-value pairs from a map
func createKVPairs(m map[string]interface{}) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}
	return b.String()
}
