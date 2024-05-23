package v3

import (
	"errors"
	"fmt"
	"time"

	brewv1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	"github.com/act3-ai/hops/internal/platform"
	"github.com/act3-ai/hops/internal/utils"
)

// Formula represents a formula's metadata.
type Formula struct {
	Desc     string `json:"desc"`
	License  string `json:"license"`
	Homepage string `json:"homepage"`
	URLs     map[string]struct {
		URL      string `json:"url"`
		Revision string `json:"revision,omitempty"`
		Tag      string `json:"tag,omitempty"`
		Branch   string `json:"branch,omitempty"`
		Using    string `json:"using,omitempty"`
		Checksum string `json:"checksum,omitempty"`
	} `json:"urls"`
	PostInstallDefined bool                               `json:"post_install_defined"`
	RubySourcePath     string                             `json:"ruby_source_path"`
	RubySourceSHA256   string                             `json:"ruby_source_sha256"`
	LinkOverwrite      []string                           `json:"link_overwrite,omitempty"`
	Revision           int                                `json:"revision,omitempty"`
	KegOnlyReason      brewv1.KegOnlyConfig               `json:"keg_only_reason,omitempty"`
	PourBottleOnlyIf   string                             `json:"pour_bottle_only_if,omitempty"`
	Caveats            string                             `json:"caveats,omitempty"`
	Service            Service                            `json:"service,omitempty"`
	VersionScheme      int                                `json:"version_scheme,omitempty"`
	Version            string                             `json:"version"`
	Bottle             Bottle                             `json:"bottle"`
	VersionedFormulae  []string                           `json:"versioned_formulae,omitempty"`
	DeprecationDate    string                             `json:"deprecation_date,omitempty"`
	DeprecationReason  string                             `json:"deprecation_reason,omitempty"`
	DisabledDate       string                             `json:"disable_date,omitempty"`
	DisabledReason     string                             `json:"disable_reason,omitempty"`
	Dependencies       Dependencies                       `json:"dependencies,omitempty"`
	HeadDependencies   Dependencies                       `json:"head_dependencies,omitempty"`
	Requirements       []Requirement                      `json:"requirements,omitempty"`
	Variations         map[platform.Platform]Dependencies `json:"variations,omitempty"`
	Conflicts          Conflicts                          `json:",inline"`
}

// Variation represents a platform-specific variation to the formula's metadata.
type Variation struct {
	// v3 caveats can only be set to a string that overwrites the general caveats.
	Caveats          string        `json:"caveats,omitempty"`
	Dependencies     Dependencies  `json:"dependencies,omitempty"`
	HeadDependencies Dependencies  `json:"head_dependencies,omitempty"`
	Requirements     []Requirement `json:"requirements,omitempty"`
	Conflicts        Conflicts     `json:",inline"`
}

// Bottle represents the bottle section.
type Bottle struct {
	Rebuild int                              `json:"rebuild"`
	RootURL string                           `json:"root_url"`
	Files   map[platform.Platform]BottleFile `json:"files"`
}

// BottleFile defines a bottle.files entry.
type BottleFile struct {
	Cellar string `json:"cellar"`
	Sha256 string `json:"sha256"`
}

// Dependencies represents a collection of dependencies.
type Dependencies map[string]*DependencyConfig

// DependencyConfig provides additional context for a dependency.
type DependencyConfig struct {
	Tags          []string     `json:"tags,omitempty"`
	UsesFromMacOS *MacOSBounds `json:"uses_from_macos,omitempty"`
}

// MacOSBounds constrains a macOS dependency.
type MacOSBounds struct {
	Since string `json:"since,omitempty"`
}

// Requirement represents a requirement.
type Requirement struct {
	Name     string   `json:"name"`
	Cask     any      `json:"cask"`
	Download any      `json:"download"`
	Version  *string  `json:"version"`
	Contexts []string `json:"contexts"`
	Specs    []string `json:"specs"`
}

// Conflicts specifies formula conflicts.
type Conflicts struct {
	ConflictsWith        []string `json:"conflicts_with,omitempty"`
	ConflictsWithReasons []string `json:"conflicts_with_reasons,omitempty"`
}

// ServiceRunType represents a run type for a service.
type ServiceRunType string

const (
	RunTypeImmediate ServiceRunType = "immediate" // immediate run type
	RunTypeInterval  ServiceRunType = "interval"  // interval run type
	RunTypeCron      ServiceRunType = "cron"      // cron run type
)

// ProcessType represents a service's process type.
type ProcessType string

const (
	ProcessTypeInteractive = "interactive" // value for interactive process
	ProcessTypeBackground  = "background"  // value for background process
)

// Service represents the service block.
//
// https://docs.brew.sh/Formula-Cookbook#service-block-methods
type Service struct {
	// Specifies per-OS service name
	// 2 examples: xinit, dbus
	// Both only specify a name for "macos"
	Name map[string]string `json:"name,omitempty"`

	// Run specifies the command to run
	//
	// One of three options:
	// - string: single argument (60 examples)
	// - []string: list of args (234 examples)
	// - map[string][]string: map OS to list of args (3 examples)
	//
	// Parse as any, require furthing parsing if needed
	Run any `json:"run"`

	RunType              ServiceRunType    `json:"run_type"`
	EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`
	Interval             time.Duration     `json:"interval,omitempty"`
	Cron                 string            `json:"cron,omitempty"`
	RequireRoot          bool              `json:"require_root,omitempty"`
	KeepAlive            struct {
		Always         bool `json:"always,omitempty"`
		SuccessfulExit bool `json:"successful_exit,omitempty"`
		Crashed        bool `json:"crashed,omitempty"` // 1 example
	} `json:"keep_alive,omitempty"`
	WorkingDir string `json:"working_dir,omitempty"`

	// 2 examples: knot-resolver, knot
	InputPath string `json:"input_path,omitempty"`

	LogPath string `json:"log_path,omitempty"`

	ErrorLogPath string `json:"error_log_path,omitempty"`

	// 3 examples: bitlbee, launch_socket_server, launchdns
	Sockets string `json:"sockets,omitempty"`

	// 4 examples
	// seen: background|interactive
	ProcessType ProcessType `json:"process_type,omitempty"`

	// 1 example: gitlab-runner
	MacOSLegacyTimers bool `json:"macos_legacy_timers,omitempty"`
}

// RunArgs evaluates the command arguments to run
// The run field is a dynamic type so it needs extra parsing.
func (s *Service) RunArgs(os string) ([]string, error) {
	switch v := s.Run.(type) {
	case string:
		return []string{v}, nil
	case []any:
		ss, err := utils.AssertStrings(v)
		if err != nil {
			return nil, fmt.Errorf("could not evaluate service command: %w", err)
		}
		return ss, nil
	case map[string]any:
		vv, ok := v[os]
		if !ok {
			return nil, nil
		}

		args, ok := vv.([]any)
		if !ok {
			return nil, errors.New("could not evaluate service command: expected list value")
		}

		ss, err := utils.AssertStrings(args)
		if err != nil {
			return nil, fmt.Errorf("could not evaluate service command: %w", err)
		}

		return ss, nil
	default:
		return nil, fmt.Errorf("could not evaluate service command: expected string, list, or map, got %T", s.Run)
	}
}
