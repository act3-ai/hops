package common

// CaskArtifact represents an artifact stanza defined by a cask.
//
// https://docs.brew.sh/Cask-Cookbook#stanza-descriptions
type CaskArtifact map[string][]any

// CaskDependencies defines Cask dependencies.
type CaskDependencies map[string][]string
