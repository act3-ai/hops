package env

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/act3-ai/hops/internal/utils/logutil"
)

// Common errors for environment variable loading
var (
	// ErrEnvVarNotFound is returned when an environment variable is not found (os.LookupEnv error)
	ErrEnvVarNotFound = errors.New("environment variable not found")
	// ErrParseEnvVar is returned when an environment variable is found but cannot be parsed
	ErrParseEnvVar = errors.New("error parsing environment variable")
)

// Or grabs the env variable or the default value.
// If there is a parsing error, it is logged with slog.Debug.
func Or[T any](name string, def T, parse func(envVal string) (T, error)) T {
	parsedVal, err := StrictOr(name, def, parse)
	if err != nil {
		// Log the parsing error
		slog.Debug("parsing env", logutil.ErrAttr(err))
		return def
	}
	return parsedVal
}

// StrictOr grabs the env variable or the default value.
// If there is a parsing error, it is returned.
func StrictOr[T any](name string, def T, parse func(envVal string) (T, error)) (T, error) {
	parsedVal, err := Required(name, parse)
	if errors.Is(err, ErrEnvVarNotFound) {
		// Eat the "environment variable not found" error
		return def, nil
	}
	// Otherwise return the values as-is
	return parsedVal, err
}

// Required grabs a required env variable.
// If the variable is not found, an error is returned.
// If there is a parsing error, it is returned.
func Required[T any](name string, parse func(envVal string) (T, error)) (T, error) {
	if name == "" {
		panic("name must not be empty")
	}
	var parsedVal T
	envVal, ok := os.LookupEnv(name)
	if !ok {
		return parsedVal, fmt.Errorf("%w: %q", ErrEnvVarNotFound, name)
	}
	parsedVal, err := parse(envVal)
	if err != nil {
		return parsedVal, fmt.Errorf("%w %q: %w", ErrParseEnvVar, name, err)
	}
	return parsedVal, nil
}

// OneOfOr grabs the first env variable found or the default value.
// If there is a parsing error, it is logged with slog.Debug.
func OneOfOr[T any](names []string, def T, parse func(envVal string) (T, error)) T {
	parsedVal, err := StrictOneOfOr(names, def, parse)
	if err != nil {
		// Log the parsing error
		slog.Debug("parsing env", logutil.ErrAttr(err))
		return def
	}
	return parsedVal
}

// StrictOr grabs the first env variable found or the default value.
// If there is a parsing error, it is returned.
func StrictOneOfOr[T any](names []string, def T, parse func(envVal string) (T, error)) (T, error) {
	parsedVal, err := OneOfRequired(names, parse)
	if errors.Is(err, ErrEnvVarNotFound) {
		// Eat the "environment variable not found" error
		return def, nil
	}
	// Otherwise return the values as-is
	return parsedVal, err
}

// OneOfRequired grabs the first env variable found.
// If none of the variables are found, an error is returned.
// If any of the parsed variables have a parsing error, it is returned.
func OneOfRequired[T any](names []string, parse func(envVal string) (T, error)) (T, error) {
	for _, name := range names {
		parsedVal, err := Required(name, parse)
		if err == nil {
			return parsedVal, nil
		}
		// Return parse errors
		if errors.Is(err, ErrParseEnvVar) {
			return parsedVal, err
		}
	}
	var emptyVal T
	return emptyVal, fmt.Errorf("%w: one of %s", ErrEnvVarNotFound, strings.Join(names, ", "))
}

// Slice grabs the env variable as a slice or the default value.
// If there is a parsing error, it is logged with slog.Debug.
func Slice[T any](name string, def []T, sep string, parse func(envVal string) (T, error)) []T {
	if name == "" {
		panic("name must not be empty")
	}
	envVal, ok := os.LookupEnv(name)
	if !ok || envVal == "" {
		return def
	}
	envVals := strings.Split(envVal, sep)
	parsedVals := make([]T, 0, len(envVals))
	for _, envVal := range envVals {
		parsedVal, err := parse(envVal)
		if err != nil {
			slog.Debug("parsing env", slog.String("val", envVal), logutil.ErrAttr(err))
		}
		parsedVals = append(parsedVals, parsedVal)
	}
	return parsedVals
}

// StrictSlice grabs the env variable as a slice or the default value.
// If there is a parsing error, it is logged with slog.Debug.
func StrictSlice[T any](name string, def []T, sep string, parse func(envVal string) (T, error)) ([]T, error) {
	parsedVals, err := RequiredSlice(name, sep, parse)
	if errors.Is(err, ErrEnvVarNotFound) {
		// Eat the "environment variable not found" error
		return def, nil
	}
	// Otherwise return the values as-is
	return parsedVals, err
}

// StrictSlice grabs the env variable as a slice or the default value.
// If there is a parsing error, it is logged with slog.Debug.
func RequiredSlice[T any](name string, sep string, parse func(envVal string) (T, error)) ([]T, error) {
	if name == "" {
		panic("name must not be empty")
	}
	envVal, ok := os.LookupEnv(name)
	if !ok || envVal == "" {
		return nil, fmt.Errorf("%w: %q", ErrEnvVarNotFound, name)
	}
	envVals := strings.Split(envVal, sep)
	parsedVals := make([]T, 0, len(envVals))
	var parseErrors error
	for _, envVal := range envVals {
		parsedVal, err := parse(envVal)
		if err != nil {
			// Join all parse errors
			parsedVals = append(parsedVals, parsedVal)
			continue
		}
		parseErrors = errors.Join(parseErrors, err)
	}
	return parsedVals, parseErrors
}
