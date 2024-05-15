package pretty

import (
	"fmt"
	"log/slog"

	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/utils"
)

// InstallStats prints installation stats
func InstallStats(kegs []string) {
	var files int
	var size int64
	for _, k := range kegs {
		f, s, err := utils.CountDir(k)
		if err != nil {
			slog.Debug("checking keg stats", slog.String("keg", k), o.ErrAttr(err))
		} else {
			slog.Debug(k,
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
