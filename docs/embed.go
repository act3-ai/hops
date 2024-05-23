package docs

import (
	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/go-common/pkg/embedutil"
)

// var GeneralDocumentation embed.FS

// Layout of embedded documentation to surface in the help command
// and generate in the gendocs command.
func Embedded(root *cobra.Command) *embedutil.Documentation {
	return &embedutil.Documentation{
		Title:   "hops",
		Command: root,
		// Categories: []*embedutil.Category{
		// 	embedutil.NewCategory(
		// 		"docs", "General Documentation", root.Name(), 1,
		// 		embedutil.LoadMarkdown(
		// 			"quick-start-guide",
		// 			"Quick Start Guide",
		// 			"quick-start-guide.md",
		// 			GeneralDocumentation),
		// 		embedutil.LoadMarkdown(
		// 			"user-guide",
		// 			"User Guide",
		// 			"user-guide.md",
		// 			GeneralDocumentation),
		// 		embedutil.LoadMarkdown(
		// 			"troubleshooting-faq",
		// 			"Troubleshooting & FAQ",
		// 			"troubleshooting-faq.md",
		// 			GeneralDocumentation),
		// 	),
		// },
	}
}
