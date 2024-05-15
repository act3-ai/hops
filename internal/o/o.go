// Package o defines functions for human-readable log messages. The package matches [Homebrew's messaging functions].
//
// [Homebrew's messaging functions]: https://docs.brew.sh/Formula-Cookbook#messaging
package o

import (
	"strings"
)

// Noop is a noop log function
func Noop(_ string) {}

// H1 prints a message, prefixed with a green arrow
//
// It is an implementation of Homebrew's Formula.oh1 function
func H1(msg string) {
	_, err := output.WriteString(styleGreen.Styled(Arrow) + " " + boldFirstLine(msg) + "\n")
	if err != nil {
		panic(err)
	}
}

// Hai prints a message, prefixed with a blue arrow
//
// It is an implementation of Homebrew's Formula.ohai function
func Hai(msg string) {
	_, err := output.WriteString(styleBlue.Styled(Arrow) + " " + boldFirstLine(msg) + "\n")
	if err != nil {
		panic(err)
	}
}

// Poo prints a warning, prefixed with "Warning:"
//
// It is an implementation of Homebrew's Formula.opoo function
// stderr
func Poo(msg string) {
	_, err := output.WriteString(styleYellow.Styled("Warning:") + " " + boldFirstLine(msg) + "\n")
	if err != nil {
		panic(err)
	}
}

// Noe prints an error, prefixed with "Error:"
//
// It is an implementation of Homebrew's Formula.onoe function
// stderr
func Noe(msg string) {
	_, err := output.WriteString(styleRed.Styled("Error: ") + " " + boldFirstLine(msg) + "\n")
	if err != nil {
		panic(err)
	}
}

// Die prints an error, prefixed with "Error:"
//
// It is an implementation of Homebrew's Formula.odie function, but does not exit like Formulae.odie
// stderr
func Die(msg string) {
	Noe(msg)
}

// Debug prints a debug message, prefixed with a magenta arrow
//
// It is an implementation of Homebrew's Formula.odebug function
func Debug(msg string) {
	_, err := output.WriteString(styleMagenta.Styled(Arrow) + " " + boldFirstLine(msg) + "\n")
	if err != nil {
		panic(err)
	}
}

func boldFirstLine(msg string) string {
	first, rest := splitFirstLine(msg)
	out := styleBold.Styled(first)
	if rest != "" {
		out += "\n" + rest
	}
	return out
}

func splitFirstLine(msg string) (string, string) {
	lines := strings.Split(msg, "\n")
	if len(lines) <= 1 {
		return msg, ""
	}
	return lines[0], strings.Join(lines[1:], "\n")
}
