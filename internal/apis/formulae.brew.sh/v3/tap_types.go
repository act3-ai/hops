package v3

// Tap represents a tap's metadata.
type Tap struct {
	TapGitHead    string             `json:"tap_git_head"`
	Aliases       map[string]string  `json:"aliases"`
	Renames       map[string]string  `json:"renames"`
	TapMigrations map[string]string  `json:"tap_migrations"`
	Formulae      map[string]Formula `json:"formulae,omitempty"`
	Casks         map[string]Cask    `json:"casks,omitempty"`
}
