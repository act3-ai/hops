package common

import (
	"errors"
	"fmt"
	"time"

	"github.com/act3-ai/hops/internal/utils"
)

// ServiceRunType represents a run type for a service.
type ServiceRunType string

// Known run types.
const (
	RunTypeImmediate ServiceRunType = "immediate" // immediate run type
	RunTypeInterval  ServiceRunType = "interval"  // interval run type
	RunTypeCron      ServiceRunType = "cron"      // cron run type
)

// ProcessType represents a service's process type.
type ProcessType string

// Known process types.
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
