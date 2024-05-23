package pretty

import (
	"fmt"
	"log/slog"

	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/prefix"
	"github.com/act3-ai/hops/internal/utils"
)

// InstallStats prints installation stats.
func InstallStats(kegs []string) {
	var files int
	var size int64
	for _, k := range kegs {
		l := slog.Default().With(slog.String("keg", k))
		if k == "" {
			l.Warn("Empty keg")
			continue
		}
		f, s, err := utils.CountDir(k)
		if err != nil {
			l.Debug("checking keg stats", o.ErrAttr(err))
		} else {
			l.Debug("Installed keg",
				slog.Int("files", f),
				slog.String("size", utils.PrettyBytes(s)),
			)
		}
		files += f
		size += s
	}
	msg := fmt.Sprintf("New kegs: %d files, %s", files, utils.PrettyBytes(size))
	fmt.Println(o.EmojiPrefixed(msg))
}

// FormulaInstallStats prints installation stats.
func FormulaInstallStats[T formula.Formula](p prefix.Prefix, formulae []T) {
	kegs := make([]string, 0, len(formulae))
	for _, f := range formulae {
		kegs = append(kegs, p.FormulaKegPath(f))
	}
	InstallStats(kegs)
}
