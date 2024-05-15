package pretty

import (
	"fmt"
	"os"
	"strings"

	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/prefix"
)

// Caveats prints formula caveats
func Caveats(f *formula.Formula, p prefix.Prefix) string {
	lines := []string{}

	if f.Caveats != nil {
		lines = append(lines, *f.Caveats)
	}

	if f.KegOnly {
		komsg := fmt.Sprintf("\n%s is keg-only, which means it was not symlinked into %s", f.Name, p)
		switch {
		case f.KegOnlyReason.Explanation != "":
			komsg += ",\nbecause " + f.KegOnlyReason.Explanation
		case f.KegOnlyReason.Reason != "":
			komsg += ",\nbecause " + string(f.KegOnlyReason.Reason)
		default:
			komsg += "."
		}
		lines = append(lines, komsg)
	}

	if f.Service != nil {
		args, err := f.Service.RunArgs("macos")
		if err != nil {
			o.Die("could not parse run args: " + err.Error())
		} else if len(args) > 0 {
			lines = append(lines,
				"",
				fmt.Sprintf("To start %s now and restart at login:", f.Name),
				"  brew services start "+f.Name,
			)
			for i, arg := range args {
				args[i] = os.ExpandEnv(arg)
			}
			lines = append(lines,
				"Or, if you don't want/need a background service you can just run:",
				"  "+strings.Join(args, " "),
			)
		}
	}

	if len(lines) == 0 {
		return ""
	}

	return strings.TrimSuffix(strings.Join(lines, "\n"), "\n")
}
