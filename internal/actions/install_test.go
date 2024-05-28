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
				Prefix: filepath.Join(tmp, "HOMEBREW_PREFIX"),
				Cache:  filepath.Join(tmp, "HOPS_CACHE"),
				Homebrew: brewenv.Configuration{
					Cache: filepath.Join(tmp, "HOMEBREW_CACHE"),
					API: brewenv.APIConfig{
						Domain: brewenv.DefaultAPIDomain,
						AutoUpdate: brewenv.AutoUpdateConfig{
							Disabled: true,
						},
					},
				},
				Registry: hopsv1.RegistryConfig{
					Prefix: "ghcr.io/homebrew/core",
				},
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
