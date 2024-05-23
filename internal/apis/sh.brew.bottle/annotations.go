package brewannotations

import (
	tab "github.com/act3-ai/hops/internal/apis/sh.brew.tab"
)

const (
	// AnnotationGitHubPackageType is the annotation key for the GitHub packages annotation.
	AnnotationGitHubPackageType = "com.github.package.type"

	// AnnotationBottleCPUVariant is the annotation key for the bottle's CPU variant.
	AnnotationBottleCPUVariant = "sh.brew.bottle.cpu.variant"

	// AnnotationBottleDigest is the annotation key for the bottle's digest.
	AnnotationBottleDigest = "sh.brew.bottle.digest"

	// AnnotationBottleDigest is the annotation key for the bottle's desired glibc version.
	AnnotationBottleGlibcVersion = "sh.brew.bottle.glibc.version"

	// AnnotationBottleSize is the annotation key for the bottle's size.
	AnnotationBottleSize = "sh.brew.bottle.size"

	// AnnotationTab is the annotation key for the bottle Tab.
	AnnotationTab = tab.AnnotationTab
)

const (
	// GitHubPackageTypeHomebrewBottle is the "com.github.package.type" value for Homebrew bottles.
	GitHubPackageTypeHomebrewBottle = "homebrew_bottle"
)
