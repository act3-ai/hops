package logutil

import (
	"context"
	"log/slog"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"

	"gitlab.com/act3-ai/asce/go-common/pkg/version"

	"github.com/act3-ai/hops/internal/utils"
)

const (
	LevelTrace   = slog.LevelDebug * 2 // trace = double debug
	LevelVerbose = slog.LevelInfo - 1  // verbose = one step more verbose than info
)

// ErrKey is the key used for errors in [log/slog] attributes.
const ErrKey = "err"

// ErrAttr produces a [log/slog.Attr] for errors.
func ErrAttr(err error) slog.Attr {
	return slog.Any(ErrKey, err)
}

// OCIPlatformValue formats an [ocispec.Platform] as a [log/slog.Attr].
func OCIPlatformValue(plat *ocispec.Platform) slog.Attr {
	if plat == nil {
		return slog.String("platform", "nil")
	}
	return slog.Attr{
		Key:   "platform",
		Value: slog.GroupValue(ociPlatformAttrs(*plat)...),
	}
}

// DescriptorGroup formats an [ocispec.Descriptor] as a [log/slog.Attr].
func DescriptorGroup(desc ocispec.Descriptor) slog.Attr {
	return slog.Attr{
		Key:   "desc",
		Value: slog.GroupValue(descriptorAttrs(desc)...),
	}
}

func ociPlatformAttrs(plat ocispec.Platform) []slog.Attr {
	return []slog.Attr{
		slog.String("architecture", plat.Architecture),
		slog.String("os", plat.OS),
		slog.String("os.version", plat.OSVersion),
	}
}

// DescriptorAttrs formats a descriptor as a list of attributes.
func DescriptorAttrs(desc ocispec.Descriptor) []any {
	return utils.ToAny(descriptorAttrs(desc))
}

// descriptorAttrs formats a descriptor as a list of attributes.
func descriptorAttrs(desc ocispec.Descriptor) []slog.Attr {
	attrs := []slog.Attr{
		slog.String("mediaType", desc.MediaType),
		slog.String("digest", desc.Digest.String()),
		slog.Int64("size", desc.Size),
	}
	if desc.ArtifactType != "" {
		attrs = append(attrs, slog.String("artifactType", desc.ArtifactType))
	}
	if desc.Annotations != nil {
		if v, ok := desc.Annotations[ocispec.AnnotationTitle]; ok {
			attrs = append(attrs, slog.String("annotations."+ocispec.AnnotationTitle, v))
		}
	}
	if desc.Platform != nil {
		attrs = append(attrs, OCIPlatformValue(desc.Platform))
	}
	return attrs
}

// WithLogging adds logging at level for the OnCopySkipped, PostCopy, and OnMounted functions.
func WithLogging(logger *slog.Logger, level slog.Level, opts *oras.CopyGraphOptions) oras.CopyGraphOptions {
	dolog := func(ctx context.Context, msg string, desc ocispec.Descriptor) {
		logger.Log(ctx, level, msg, //nolint:sloglint
			DescriptorAttrs(desc)...,
		)
	}

	onCopySkipped := opts.OnCopySkipped
	opts.OnCopySkipped = func(ctx context.Context, desc ocispec.Descriptor) error {
		dolog(ctx, "Skipped copy for artifact", desc)

		if onCopySkipped != nil {
			return onCopySkipped(ctx, desc)
		}
		return nil
	}
	postCopy := opts.PostCopy
	opts.PostCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
		dolog(ctx, "Copied artifact", desc)
		if postCopy != nil {
			return postCopy(ctx, desc)
		}
		return nil
	}
	onMounted := opts.OnMounted
	opts.OnMounted = func(ctx context.Context, desc ocispec.Descriptor) error {
		dolog(ctx, "Mounted artifact", desc)
		if onMounted != nil {
			return onMounted(ctx, desc)
		}
		return nil
	}
	return *opts
}

// VersionAttrs formats version info as a list of [log/slog.Attr].
func VersionAttrs(info version.Info) []slog.Attr {
	attrs := []slog.Attr{
		slog.String("version", info.Version),
		slog.String("commit", info.Commit),
		slog.Bool("dirty", info.Dirty),
		slog.String("built", info.Built),
	}
	return attrs
}
