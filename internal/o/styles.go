package o

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"

	"github.com/act3-ai/hops/internal/utils"
	hlog "github.com/act3-ai/hops/internal/utils/logutil"
)

const (
	Arrow   = "==>" // arrow used to prefix log message headers
	Check   = "✔"   // Check mark used to indicate success
	Warning = "⚠"   // Exclamation sign used to indicate warning
	X       = "✘"   // "X" mark used to indicate failure
)

// Defined styles
var (
	// Returns width of terminal clamped between 80 and 120
	// With of 80 is used for non-TTY terminals
	Width = min(utils.TerminalWidth(80), 120)

	output  = termenv.DefaultOutput() // Terminal output profile
	red     = termenv.ANSIRed
	yellow  = termenv.ANSIYellow
	green   = termenv.ANSIGreen
	blue    = termenv.ANSIBlue
	magenta = termenv.ANSIMagenta

	styleBold      = output.String().Bold()
	styleUnderline = output.String().Underline()
	styleFaint     = output.String().Faint()
	styleRed       = output.String().Foreground(red)
	styleYellow    = output.String().Foreground(yellow)
	styleGreen     = output.String().Foreground(green)
	styleBlue      = output.String().Foreground(blue)
	styleMagenta   = output.String().Foreground(magenta)
)

// Semantic aliases for defined styles
var (
	StyleSuccess = StyleGreen  // semantic name for the green style
	StyleWarning = StyleYellow // semantic name for the yellow style
	StyleError   = StyleRed    // semantic name for the red style
)

// Styles each string
func StyleEach(style termenv.Style, s []string) []string {
	styled := make([]string, len(s))
	copy(styled, s)
	for i, ss := range styled {
		styled[i] = style.Styled(ss)
	}
	return styled
}

// Styles string bold
func StyleBold(s string) string {
	return styleBold.Styled(s)
}

// Styles string faded
func StyleFaint(s string) string {
	return styleFaint.Styled(s)
}

// Styles string underlined
func StyleUnderline(s string) string {
	return styleUnderline.Styled(s)
}

// Styles string blue
func StyleBlue(s string) string {
	return styleBlue.Styled(s)
}

// Styles string green
func StyleGreen(s string) string {
	return styleGreen.Styled(s)
}

// Styles string red
func StyleRed(s string) string {
	return styleRed.Styled(s)
}

// Styles string yellow
func StyleYellow(s string) string {
	return styleYellow.Styled(s)
}

// Returns log styles for charmbracelet/log
func LogStyles() *log.Styles {
	styles := log.DefaultStyles()

	styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().
		SetString("Error:").
		Foreground(lipgloss.ANSIColor(red))
	styles.Levels[log.WarnLevel] = lipgloss.NewStyle().
		SetString("Warning:").
		Foreground(lipgloss.ANSIColor(yellow))
	styles.Levels[log.InfoLevel] = lipgloss.NewStyle().SetString(Arrow).Foreground(lipgloss.ANSIColor(blue))
	styles.Levels[log.DebugLevel] = lipgloss.NewStyle().
		SetString(Arrow).
		Foreground(lipgloss.ANSIColor(magenta))

	// Add a custom style for key err/error
	styles.Keys[ErrKey] = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(red))

	return styles
}

// aliases for internal log utility (avoid import cycles)
var (
	ErrKey  = hlog.ErrKey
	ErrAttr = hlog.ErrAttr
)
