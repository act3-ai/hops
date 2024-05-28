package o

import (
	"github.com/muesli/termenv"
)

var (
	noEmoji      = false                 // initial setting for emoji
	installBadge = "🌼"                   // initial setting for badge
	color        = !termenv.EnvNoColor() // initial setting for color
)

// EmojiPrefixed returns msg prefixed with the install badge
// Obeys the NoEmoji and InstallBadge settings.
func EmojiPrefixed(msg string) string {
	if NoEmoji() {
		return msg
	}
	return InstallBadge() + " " + msg
}

// NoEmoji reports the NoEmoji setting.
func NoEmoji() bool {
	return noEmoji
}

// InstallBadge returns the install badge.
func InstallBadge() string {
	return installBadge
}

// Color reports the color setting.
func Color() bool {
	return color
}

// SetNoEmoji sets the NoEmoji setting.
func SetNoEmoji(value bool) {
	noEmoji = value
}

// SetInstallBadge sets the install badge setting.
func SetInstallBadge(value string) {
	installBadge = value
}
