package o

// PrettyInstalled formats an installed formula's name.
func PrettyInstalled(formula string) string {
	if NoEmoji() {
		return StyleSuccess(StyleBold(formula + " (installed)"))
	}
	return StyleBold(formula + " " + StyleSuccess(Check))
}

// PrettyOutdated formats an outdated formula's name.
func PrettyOutdated(formula string) string {
	if NoEmoji() {
		return StyleError(StyleBold(formula + " (outdated)"))
	}
	return StyleBold(formula + " " + StyleError(Warning))
}

// PrettyUninstalled formats an uninstalled formula's name.
func PrettyUninstalled(formula string) string {
	if NoEmoji() {
		return StyleError(StyleBold(formula + " (uninstalled)"))
	}
	return StyleBold(formula + " " + StyleError(X))
}
