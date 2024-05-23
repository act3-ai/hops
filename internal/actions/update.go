package actions

import (
	"context"
	"log/slog"
	"strings"

	"golang.org/x/mod/semver"

	hopsv1 "github.com/act3-ai/hops/internal/apis/config.hops.io/v1beta1"
	brewv1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	brewapi "github.com/act3-ai/hops/internal/brew/api"
	brewformulary "github.com/act3-ai/hops/internal/brew/formulary"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/utils"
)

// Update represents the action and its options.
type Update struct {
	*Hops
}

// Run runs the action.
func (action *Update) Run(ctx context.Context) error {
	if action.Config().Registry.Prefix != "" {
		o.Hai("Update not necessary for standalone registry mode")
		return nil
	}

	apiclient := brewapi.NewClient(action.Config().Homebrew.Domain)

	// Only load the cached indexes
	oldIndex, err := brewformulary.LoadV1(action.Config().Cache)
	if err != nil {
		slog.Warn("loading cached index", o.ErrAttr(err))
	}

	newIndex, err := brewformulary.FetchV1(ctx,
		apiclient,
		action.Config().Cache,
		&hopsv1.AutoUpdateConfig{
			Secs: new(int), // set refresh seconds to zero
		})
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

// IsNewerThan reports if n is newer than o by comparing their versions.
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
