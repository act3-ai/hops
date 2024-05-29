package formula

import (
	"github.com/act3-ai/hops/internal/apis/formulae.brew.sh/common"
	brewv1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	"github.com/act3-ai/hops/internal/platform"
)

// V1 implements Formula for v1 API output.
type V1 struct {
	src     brewv1.Info
	version Version
	info    *Info
}

// SourceV1 returns the underlying v1 API output.
func (f *V1) SourceV1() *brewv1.Info {
	return &f.src
}

// Name implements Formula.
func (f *V1) Name() string {
	return f.src.Name
}

// Version implements Formula.
func (f *V1) Version() Version {
	return f.version
}

// Info implements Formula.
func (f *V1) Info() *Info {
	return f.info
}

// platformFormulaV1 implements PlatformFormula for v1 API output.
type platformFormulaV1 struct {
	src       brewv1.PlatformInfo
	plat      platform.Platform
	version   Version
	info      *Info
	conflicts []Conflict
	bottle    *Bottle
}

// Bottle implements PlatformFormula.
func (p *platformFormulaV1) Bottle() *Bottle {
	return p.bottle
}

// Service implements PlatformFormulaWithInfo.
func (p *platformFormulaV1) Service() *common.Service {
	return p.src.Service
}

// LinkOverwrite implements PlatformFormulaWithInfo.
func (p *platformFormulaV1) LinkOverwrite() []string {
	return p.src.LinkOverwrite
}

// IsKegOnly implements PlatformFormulaWithInfo.
func (p *platformFormulaV1) IsKegOnly() bool {
	return p.src.KegOnly
}

// KegOnlyReason implements PlatformFormulaWithInfo.
func (p *platformFormulaV1) KegOnlyReason() (reason string) {
	komsg := string(p.src.KegOnlyReason.Reason)
	if p.src.KegOnlyReason.Explanation != "" {
		komsg = p.src.KegOnlyReason.Explanation
	}
	return komsg
}

// Caveats implements PlatformFormulaWithInfo.
func (p *platformFormulaV1) Caveats() string {
	if p.src.Caveats != nil {
		return *p.src.Caveats
	}
	return ""
}

// Conflicts implements PlatformFormulaWithInfo.
func (p *platformFormulaV1) Conflicts() []Conflict {
	if p.conflicts == nil {
		p.conflicts = make([]Conflict, 0, len(p.src.ConflictsWith))
		for i, with := range p.src.ConflictsWith {
			p.conflicts = append(p.conflicts, Conflict{
				Name:   with,
				Reason: p.src.ConflictsWithReasons[i],
			})
		}
	}
	return p.conflicts
}

// Dependencies implements PlatformFormulaWithInfo.
func (p *platformFormulaV1) Dependencies() *TaggedDependencies {
	return &TaggedDependencies{
		Required:    p.src.Dependencies,
		Build:       p.src.BuildDependencies,
		Test:        p.src.TestDependencies,
		Recommended: p.src.RecommendedDependencies,
		Optional:    p.src.OptionalDependencies,
	}
}

// Info implements PlatformFormulaWithInfo.
func (p *platformFormulaV1) Info() *Info {
	return p.info
}

// SystemDependencies implements PlatformFormulaWithInfo.
func (*platformFormulaV1) SystemDependencies() *TaggedDependencies {
	// return &taggedDependencies{
	// 	required:    p.src.Dependencies,
	// 	build:       p.src.BuildDependencies,
	// 	test:        p.src.TestDependencies,
	// 	recommended: p.src.RecommendedDependencies,
	// 	optional:    p.src.OptionalDependencies,
	// }
	// TODO: parse uses_from_macos entries
	return &TaggedDependencies{}
}

// SourceInfo implements PlatformFormula.
func (p *platformFormulaV1) SourceInfo() *SourceInfo {
	stable := p.src.URLs[brewv1.Stable]
	return &SourceInfo{
		URL:      stable.URL,
		Using:    stable.Using,
		Checksum: stable.Checksum,
		Git: GitSource{
			Revision: stable.Revision,
			Tag:      stable.Tag,
			Branch:   stable.Branch,
		},
		Ruby: RubySource{
			Path:   p.src.RubySourcePath,
			Sha256: p.src.RubySourceChecksum[brewv1.RubySourceChecksumSha256],
		},
	}
}

// Name implements PlatformFormula.
func (p *platformFormulaV1) Name() string {
	return p.src.Name
}

// Platform implements PlatformFormula.
func (p *platformFormulaV1) Platform() platform.Platform {
	return p.plat
}

// Version implements PlatformFormula.
func (p *platformFormulaV1) Version() Version {
	return p.version
}

// FromV1 produces a Formula from v1 API input.
func FromV1(input *brewv1.Info) MultiPlatformFormula {
	return &V1{
		src:     *input,
		version: versionFromV1(&input.PlatformInfo),
		info:    infoFromV1(&input.PlatformInfo),
	}
}

// ForPlatform implements MultiPlatformFormula.
func (f *V1) ForPlatform(plat platform.Platform) (PlatformFormula, error) {
	pf, err := f.src.ForPlatform(plat)
	if err != nil {
		return nil, err
	}
	return PlatformFromV1(plat, pf), nil
}

// PlatformFromV1 creates a Formula from v1 API input.
func PlatformFromV1(plat platform.Platform, input *brewv1.PlatformInfo) PlatformFormula {
	return &platformFormulaV1{
		src:     *input,
		plat:    plat,
		version: versionFromV1(input),
		info:    infoFromV1(input),
		bottle:  bottleFromV1(plat, input),
	}
}

func versionFromV1(input *brewv1.PlatformInfo) Version {
	rebuild := 0
	if stable, ok := input.Bottle[brewv1.Stable]; ok && stable != nil {
		rebuild = stable.Rebuild
	}
	return &version{
		version:  input.Versions.Stable,
		revision: input.Revision,
		rebuild:  rebuild,
	}
}

func infoFromV1(input *brewv1.PlatformInfo) *Info {
	return &Info{
		Desc:     input.Desc,
		License:  input.License,
		Homepage: input.Homepage,
	}
}

func bottleFromV1(plat platform.Platform, input *brewv1.PlatformInfo) *Bottle {
	stable, ok := input.Bottle[brewv1.Stable]
	if !ok || stable == nil {
		return nil
	}

	var fplat platform.Platform
	var file *brewv1.BottleFile

	if pfile, ok := stable.Files[plat]; ok && pfile != nil {
		fplat = plat
		file = pfile
	} else if afile, ok := stable.Files[platform.All]; ok && afile != nil {
		// Fallback to "all" Bottle
		fplat = platform.All
		file = afile
	} else {
		// No Bottle for platform
		return nil
	}

	pourOnlyIf := ""
	if input.PourBottleOnlyIf != nil {
		pourOnlyIf = *input.PourBottleOnlyIf
	}

	bottle := &Bottle{
		RootURL:    stable.RootURL,
		Sha256:     file.Sha256,
		Cellar:     file.Cellar,
		Platform:   fplat,
		PourOnlyIf: pourOnlyIf,
	}

	// Set Cellar ot empty string if relocatable
	if file.Relocatable() {
		bottle.Cellar = ""
	}

	return bottle
}
