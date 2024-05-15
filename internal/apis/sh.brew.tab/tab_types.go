package tab

import (
	v1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
)

const (
	// AnnotationTab is the annotation key for the bottle Tab
	AnnotationTab = "sh.brew.tab"
)

// Tab represents the contents of the sh.brew.tab annotation
//
// This annotation is added to each bottle manifest
type Tab struct {
	HomebrewVersion     string                  `json:"homebrew_version"`
	ChangedFiles        []string                `json:"changed_files"`
	SourceModifiedTime  uint                    `json:"source_modified_time"`
	StdLib              string                  `json:"stdlib,omitempty"`
	Compiler            string                  `json:"compiler"`
	RuntimeDependencies []*v1.RuntimeDependency `json:"runtime_dependencies"`
	Arch                string                  `json:"arch"`
	BuiltOn             BuiltOn                 `json:"built_on"`
}

// BuiltOn stores build information
type BuiltOn struct {
	OS            string `json:"os"`
	OSVersion     string `json:"os_version"`
	CPUFamily     string `json:"cpu_family"`
	XCode         string `json:"xcode"`
	CLT           string `json:"clt"`
	PreferredPerl string `json:"preferred_perl"`
}
