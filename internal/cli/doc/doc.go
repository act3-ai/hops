package doc

import (
	"fmt"
	"strings"

	"github.com/muesli/termenv"
)

// Header renders a header as a Markdown h3 if color output is disabled.
func Header(s string) string {
	if termenv.DefaultOutput().EnvNoColor() {
		// Return Markdown-formatted
		return "### " + strings.TrimSuffix(s, ":") + "\n"
	}
	return s
}

// Code renders an inline Code block as Markdown if color output is disabled.
func Code(s string) string {
	if termenv.DefaultOutput().EnvNoColor() {
		// Return Markdown-formatted
		return "`" + s + "`"
	}
	return s
}

// CodeBlock renders a code block as Markdown if color output is disabled.
func CodeBlock(language, s string) string {
	if termenv.DefaultOutput().EnvNoColor() {
		// Return Markdown-formatted
		return "\n```" + language + "\n" + strings.TrimSuffix(s, "\n") + "\n```"
	}
	return s
}

// Footer renders a footer as Markdown if color output is disabled.
func Footer(s string) string {
	if termenv.DefaultOutput().EnvNoColor() {
		// Return Markdown-formatted
		lines := []string{}
		for _, line := range strings.Split(strings.TrimSpace(s), "\n") {
			if strings.TrimSpace(line) == "" {
				lines = append(lines, ">") // no trailing space for lint reasons
			} else {
				lines = append(lines, "> "+line) // no trailing space for lint reasons
			}
		}
		return strings.Join(lines, "\n")
	}
	return s
}

// UList renders an unordered list as Markdown if color output is disabled.
func UList(defaultPrefix string, items ...string) string {
	if termenv.DefaultOutput().EnvNoColor() {
		// Return Markdown-formatted with starting newline
		result := "\n"
		for _, item := range items {
			result += "- " + item + "\n"
		}
		return strings.TrimSuffix(result, "\n")
	}

	// Return formatted with default prefix
	result := ""
	for _, item := range items {
		result += defaultPrefix + item + "\n"
	}
	return strings.TrimSuffix(result, "\n")
}

// OList renders an ordered list as Markdown if color output is disabled.
func OList(items ...string) string {
	if termenv.DefaultOutput().EnvNoColor() {
		// Return Markdown-formatted with starting newline
		result := "\n"
		for i, item := range items {
			result += fmt.Sprintf("%d. %s\n", i+1, item)
		}
		return strings.TrimSuffix(result, "\n")
	}

	// Return formatted with no leading newline
	result := ""
	for i, item := range items {
		result += fmt.Sprintf("%d. %s\n", i+1, item)
	}
	return strings.TrimSuffix(result, "\n")
}
