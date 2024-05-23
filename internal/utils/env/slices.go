package env

import (
	"strings"
)

// SpaceSlice represents a slice of strings.
type SpaceSlice []string

// Implements go-env/env.Unmarshaler.
func (s *SpaceSlice) UnmarshalEnvironmentValue(data string) error {
	*s = strings.Split(data, " ")
	return nil
}

// Implements go-env/env.Marshaler.
func (s SpaceSlice) MarshalEnvironmentValue() (string, error) {
	strings.Join(s, " ")
	return "", nil
}

// CommaSlice represents a slice of strings.
type CommaSlice []string

// Implements go-env/env.Unmarshaler.
func (s *CommaSlice) UnmarshalEnvironmentValue(data string) error {
	*s = strings.Split(data, ",")
	return nil
}

// Implements go-env/env.Marshaler.
func (s CommaSlice) MarshalEnvironmentValue() (string, error) {
	strings.Join(s, ",")
	return "", nil
}
