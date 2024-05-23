package debugutil

import (
	"encoding/json"
)

// DebugMarshalJSON is used for debug printing.
func DebugMarshalJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}
