package actions

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	brewenv "github.com/act3-ai/hops/internal/apis/config.brew.sh"
	hopsv1 "github.com/act3-ai/hops/internal/apis/config.hops.io/v1beta1"
)

func BenchmarkInstall(b *testing.B) {
	tmp := b.TempDir()

	action := &Install{
		Hops: &Hops{
			version:     "test",
			ConfigFiles: []string{},
			EnvFiles:    []string{},
			Concurrency: runtime.NumCPU(),
			cfg: &hopsv1.Configuration{
				Cache: filepath.Join(tmp, "HOPS_CACHE"),
				Homebrew: hopsv1.HomebrewAPIConfig{
					Domain: brewenv.Default.APIDomain,
					AutoUpdate: hopsv1.AutoUpdateConfig{
						Disabled: true,
					},
				},
				Registry: hopsv1.RegistryConfig{
					Prefix: "ghcr.io/homebrew/core",
				},
			},
			brewcfg: &brewenv.Environment{
				Prefix:    filepath.Join(tmp, "HOMEBREW_PREFIX"),
				APIDomain: brewenv.Default.APIDomain,
				Cache:     filepath.Join(tmp, "HOMEBREW_CACHE"),
			},
		},
	}
	for i := 0; i < b.N; i++ {
		err := action.Run(context.Background(), "gh")
		if err != nil {
			b.Fatal(err)
		}
	}
}
