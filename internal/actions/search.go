package actions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"

	brewv1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	brewapi "github.com/act3-ai/hops/internal/brew/api"
	brewformulary "github.com/act3-ai/hops/internal/brew/formulary"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/utils/logutil"
)

// Search represents the action and its options.
type Search struct {
	*Hops

	// Search for formulae with a description matching text
	// and casks with a name or description matching text
	Desc bool
}

// Run runs the action.
func (action *Search) Run(ctx context.Context, terms ...string) error {
	if action.Config().Registry.Prefix != "" {
		o.Hai("Search is not available for standalone registry mode")
		return nil
	}

	matchFuncs, err := parseTerms(terms)
	if err != nil {
		return err
	}

	// Load the index
	index, err := brewformulary.FetchV1(ctx,
		brewapi.NewClient(action.Config().Homebrew.API.Domain),
		action.Config().Cache, nil)
	if err != nil {
		return err
	}

	var hits []string
	if !action.Desc {
		for _, f := range index.ListNames() {
			for _, match := range matchFuncs {
				if match(f) {
					hits = append(hits, f)
				}
			}
		}
	} else {
		fhits := index.SearchFunc(func(f *brewv1.Info) bool {
			for _, match := range matchFuncs {
				// Check against descriptions
				if match(f.Desc) {
					return true
				}
			}
			return false
		})
		for _, f := range fhits {
			hits = append(hits, f.Name)
		}
	}

	if len(hits) == 0 {
		return fmt.Errorf("no matches found for %q", strings.Join(terms, " "))
	}

	slog.Info("Formulae")

	for _, hit := range hits {
		hitString := hit
		if isInstalled(
			action.Prefix().Cellar(), hit) {
			hitString = o.PrettyInstalled(hit)
		}
		fmt.Println(hitString)
	}

	if !allAlphanumeric(terms) {
		slog.Warn(heredoc.Doc(`
			Did you mean to perform a regular expression search?
			Surround your query with /slashes/ to search locally by regex.`))
	}

	return nil
}

// parseTerms parses a list of search terms into a list of match functions.
func parseTerms(terms []string) ([]func(s string) bool, error) {
	matchFuncs := []func(s string) bool{}
	for _, term := range terms {
		// Regex terms are indicated by wrapping the text in "/" characters
		if strings.HasPrefix(term, "/") && strings.HasSuffix(term, "/") {
			expr := strings.TrimPrefix(strings.TrimSuffix(term, "/"), "/")
			re, err := regexp.Compile(expr)
			if err != nil {
				return matchFuncs, err
			}

			// Add a function that checks for substring existence in both directions
			matchFuncs = append(matchFuncs, func(s string) bool {
				return re.MatchString(s)
			})
			continue
		}

		// Add a function that checks for substring existence in both directions
		matchFuncs = append(matchFuncs, func(s string) bool {
			return strings.Contains(s, term)
		})
	}
	return matchFuncs, nil
}

// alphanumeric is a regex expression to check if a string is alphanumeric.
// also allows for underscore and dash characters.
var alphanumeric = regexp.MustCompile("^[a-zA-Z0-9_-]*$")

// isAlphanumeric reports if a string is alphanumeric.
// also allows for underscore and dash characters.
func isAlphanumeric(s string) bool {
	return alphanumeric.MatchString(s)
}

// allAlphanumeric reports if all strings in the list are alphanumeric.
// also allows for underscore and dash characters.
func allAlphanumeric(terms []string) bool {
	for _, term := range terms {
		if !isAlphanumeric(term) {
			return false
		}
	}
	return true
}

// isInstalled reports if formula with name "name" is installed.
func isInstalled(cellar, name string) bool {
	dir := filepath.Join(cellar, name)
	entries, err := os.ReadDir(dir)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return false
	case err != nil:
		slog.Warn("checking cellar", logutil.ErrAttr(err))
		return false
	case len(entries) == 0:
		return false
	default:
		return true
	}
}
