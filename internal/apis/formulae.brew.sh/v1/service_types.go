package v1

import (
	"errors"
	"fmt"
	"time"

	"github.com/act3-ai/hops/internal/utils"
)

// ServiceRunType represents a run type for a service
type ServiceRunType string

const (
	RunTypeImmediate ServiceRunType = "immediate" // immediate run type
	RunTypeInterval  ServiceRunType = "interval"  // interval run type
	RunTypeCron      ServiceRunType = "cron"      // cron run type
)

// ProcessType represents a service's process type
type ProcessType string

const (
	ProcessTypeInteractive = "interactive" // value for interactive process
	ProcessTypeBackground  = "background"  // value for background process
)

// Service represents the service block
//
// https://docs.brew.sh/Formula-Cookbook#service-block-methods
type Service struct {
	Name *struct {
		MacOS string `json:"macos,omitempty"`
	} `json:"name,omitempty"`
	Run                  any               `json:"run"`
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
	WorkingDir        string      `json:"working_dir,omitempty"`
	InputPath         string      `json:"input_path,omitempty"` // 2 examples
	LogPath           string      `json:"log_path,omitempty"`
	ErrorLogPath      string      `json:"error_log_path,omitempty"`
	Sockets           string      `json:"sockets,omitempty"`             // 3 examples
	ProcessType       ProcessType `json:"process_type,omitempty"`        // seen: background|interactive
	MacOSLegacyTimers bool        `json:"macos_legacy_timers,omitempty"` // literally only used on the gitlab-runners formula
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
