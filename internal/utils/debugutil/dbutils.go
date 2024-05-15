package debugutil

import (
	"encoding/json"
)

// DebugMarshalJSON
func DebugMarshalJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}
