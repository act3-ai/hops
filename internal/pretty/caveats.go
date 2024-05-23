package pretty

import (
	"fmt"
	"os"
	"strings"

	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/prefix"
)

// Caveats prints formula caveats.
func Caveats(f formula.PlatformFormula, p prefix.Prefix) string {
	lines := []string{}

	if caveats := f.Caveats(); caveats != "" {
		lines = append(lines, f.Caveats())
	}

	if f.IsKegOnly() {
		komsg := fmt.Sprintf("\n%s is keg-only, which means it was not symlinked into %s", f.Name(), p)
		switch {
		case f.KegOnlyReason() != "":
			komsg += ",\nbecause " + f.KegOnlyReason()
		default:
			komsg += "."
		}
		lines = append(lines, komsg)
	}

	if service := f.Service(); service != nil {
		args, err := service.RunArgs("macos")
		if err != nil {
			o.Die("could not parse run args: " + err.Error())
		} else if len(args) > 0 {
			lines = append(lines,
				"",
				fmt.Sprintf("To start %s now and restart at login:", f.Name()),
				"  brew services start "+f.Name(),
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
