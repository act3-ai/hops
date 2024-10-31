package platform

import (
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestFromOCI(t *testing.T) {
	type args struct {
		r *ocispec.Platform
	}
	tests := []struct {
		name string
		args args
		want Platform
	}{
		{
			name: "Arm64Sequoia",
			args: args{
				r: &ocispec.Platform{
					Architecture: "arm64",
					OS:           "darwin",
					OSVersion:    "macOS 15.1",
				},
			},
			want: Arm64Sequoia,
		},
		{
			name: "Arm64Sonoma",
			args: args{
				r: &ocispec.Platform{
					Architecture: "arm64",
					OS:           "darwin",
					OSVersion:    "macOS 14.4",
				},
			},
			want: Arm64Sonoma,
		},
		{
			name: "Arm64Ventura",
			args: args{
				r: &ocispec.Platform{
					Architecture: "arm64",
					OS:           "darwin",
					OSVersion:    "macOS 13.6",
				},
			},
			want: Arm64Ventura,
		},
		{
			name: "Arm64Monterey",
			args: args{
				r: &ocispec.Platform{
					Architecture: "arm64",
					OS:           "darwin",
					OSVersion:    "macOS 12.7",
				},
			},
			want: Arm64Monterey,
		},
		{
			name: "Arm64BigSur",
			args: args{
				r: &ocispec.Platform{
					Architecture: "arm64",
					OS:           "darwin",
					OSVersion:    "macOS 11.0",
				},
			},
			want: Arm64BigSur,
		},
		{
			name: "Sequoia",
			args: args{
				r: &ocispec.Platform{
					Architecture: "arm64",
					OS:           "darwin",
					OSVersion:    "macOS 15.1",
				},
			},
			want: Sequoia,
		},
		{
			name: "Sonoma",
			args: args{
				r: &ocispec.Platform{
					Architecture: "amd64",
					OS:           "darwin",
					OSVersion:    "macOS 14.4",
				},
			},
			want: Sonoma,
		},
		{
			name: "Ventura",
			args: args{
				r: &ocispec.Platform{
					Architecture: "amd64",
					OS:           "darwin",
					OSVersion:    "macOS 13.6",
				},
			},
			want: Ventura,
		},
		{
			name: "Monterey",
			args: args{
				r: &ocispec.Platform{
					Architecture: "amd64",
					OS:           "darwin",
					OSVersion:    "macOS 12.7",
				},
			},
			want: Monterey,
		},
		{
			name: "BigSur",
			args: args{
				r: &ocispec.Platform{
					Architecture: "amd64",
					OS:           "darwin",
					OSVersion:    "macOS 11.0",
				},
			},
			want: BigSur,
		},
		{
			name: "Catalina",
			args: args{
				r: &ocispec.Platform{
					Architecture: "amd64",
					OS:           "darwin",
					OSVersion:    "macOS 10.15",
				},
			},
			want: Catalina,
		},
		{
			name: "Mojave",
			args: args{
				r: &ocispec.Platform{
					Architecture: "amd64",
					OS:           "darwin",
					OSVersion:    "macOS 10.14",
				},
			},
			want: Mojave,
		},
		{
			name: "HighSierra",
			args: args{
				r: &ocispec.Platform{
					Architecture: "amd64",
					OS:           "darwin",
					OSVersion:    "macOS 10.13",
				},
			},
			want: HighSierra,
		},
		{
			name: "X8664Linux",
			args: args{
				r: &ocispec.Platform{
					Architecture: "amd64",
					OS:           "linux",
					OSVersion:    "Ubuntu 22.04",
				},
			},
			want: X8664Linux,
		},
		{
			name: "All: empty fields",
			args: args{
				r: &ocispec.Platform{
					Architecture: "",
					OS:           "",
					OSVersion:    "",
				},
			},
			want: All,
		},
		{
			name: "Unsupported: linux/arm64",
			args: args{
				r: &ocispec.Platform{
					Architecture: "arm64",
					OS:           "linux",
					OSVersion:    "Ubuntu 22.04",
				},
			},
			want: Unsupported,
		},
		{
			name: "Unsupported: linux/arm",
			args: args{
				r: &ocispec.Platform{
					Architecture: "arm",
					OS:           "linux",
					OSVersion:    "Ubuntu 22.04",
				},
			},
			want: Unsupported,
		},
		{
			name: "Unsupported: darwin/arm",
			args: args{
				r: &ocispec.Platform{
					Architecture: "arm",
					OS:           "darwin",
					OSVersion:    "macOS 14.4",
				},
			},
			want: Unsupported,
		},
		{
			name: "Unsupported: freebsd",
			args: args{
				r: &ocispec.Platform{
					Architecture: "amd64",
					OS:           "freebsd",
				},
			},
			want: Unsupported,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromOCI(tt.args.r); got != tt.want {
				t.Errorf("FromOCI() = %v, want %v", got, tt.want)
			}
		})
	}
}
