package common

// Pour bottle conditions.
const (
	PourBottleConditionDefaultPrefix = "default_prefix" // pour bottle condition requiring the default prefix
	PourBottleConditionCLTInstalled  = "clt_installed"  // pour bottle condition requiring the macOS command line tools
)

// Bottle Cellar values.
const (
	CellarAny               = ":any"                 // Signifies bottle is safe to install in the Cellar.
	CellarAnySkipRelocation = ":any_skip_relocation" // Signifies bottle is safe to install in the Cellar without relocation.
)

// Relocatable reports if the Cellar value means the Bottle is relocatable.
func CellarRelocatable(c string) bool {
	return c == CellarAny || c == CellarAnySkipRelocation
}
