package actions

import (
	"context"
	"log/slog"
	"strings"

	"golang.org/x/mod/semver"

	brewv1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/utils"
)

// Update represents the action and its options
type Update struct {
	*Hops
}

// Run runs the action
func (action *Update) Run(ctx context.Context) error {
	oldIndex := action.Index()

	// Only load the cached indexes
	if oldIndex.IsCached() {
		err := oldIndex.Load(ctx)
		if err != nil {
			slog.Warn("loading cached index", o.ErrAttr(err))
		}
	}

	cfg := action.Config()

	newIndex := action.Index()

	// Force a reset and redownload
	err := newIndex.Reset(&cfg.Homebrew.AutoUpdate)
	if err != nil {
		return err
	}

	err = newIndex.Load(ctx)
	if err != nil {
		return err
	}

	if oldIndex != nil {
		updated := []string{}
		added := []string{}

		for _, newf := range newIndex.List() {
			oldf := oldIndex.Find(newf.Name)
			if oldf == nil {
				added = append(added, newf.Name)
				continue // remaining checks assume previous version exists
			}
			if IsNewerThan(newf, oldf) {
				updated = append(updated, newf.Name)
			}
		}

		if len(updated) > 0 {
			o.Hai("Updated formulae\n" + strings.Join(updated, "\n"))
		}

		if len(added) > 0 {
			o.Hai("New formulae\n" + strings.Join(added, "\n"))
		}
	}

	return nil
}

// IsNewerThan reports if n is newer than o by comparing their versions
func IsNewerThan(n *brewv1.Info, o *brewv1.Info) bool {
	return semver.Compare(
		utils.FmtSemver(n.Versions.Stable),
		utils.FmtSemver(o.Versions.Stable),
	) > 0
}

func versionCompare(n, o string) int {
	return semver.Compare(
		utils.FmtSemver(n),
		utils.FmtSemver(o),
	)
}
