package logutil

import (
	"context"
	"log/slog"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
)

// ErrKey is the slog attribute key used for errors in log messages
const ErrKey = "err"

// ErrAttr produces a slog.Attr for errors
func ErrAttr(err error) slog.Attr {
	return slog.Any(ErrKey, err)
}

// OCIPlatformValue formats an OCI platform for logging
func OCIPlatformValue(plat *ocispec.Platform) slog.Attr {
	if plat == nil {
		return slog.String("platform", "nil")
	}
	return slog.Attr{
		Key:   "platform",
		Value: slog.GroupValue(ociPlatformAttrs(*plat)...),
	}
}

// Descriptor formats an OCI descriptor for logging
func Descriptor(desc ocispec.Descriptor) slog.Attr {
	return slog.Attr{
		Key:   "desc",
		Value: slog.GroupValue(descriptorValues(desc)...),
	}
}

func ociPlatformAttrs(plat ocispec.Platform) []slog.Attr {
	return []slog.Attr{
		slog.String("architecture", plat.Architecture),
		slog.String("os", plat.OS),
		slog.String("os.version", plat.OSVersion),
	}
}

func descriptorValues(desc ocispec.Descriptor) []slog.Attr {
	attrs := []slog.Attr{
		slog.String("mediaType", desc.MediaType),
		slog.String("digest", desc.Digest.String()),
		slog.Int64("size", desc.Size),
	}
	if desc.Platform != nil {
		attrs = append(attrs, OCIPlatformValue(desc.Platform))
	}
	return attrs
}

// WithLogging adds logging at level for the OnCopySkipped, PostCopy, and OnMounted functions
func WithLogging(logger *slog.Logger, level slog.Level, opts *oras.CopyGraphOptions) oras.CopyGraphOptions {
	onCopySkipped := opts.OnCopySkipped
	opts.OnCopySkipped = func(ctx context.Context, desc ocispec.Descriptor) error {
		logger.Log(ctx, level, "Skipped artifact", Descriptor(desc))
		if onCopySkipped != nil {
			return onCopySkipped(ctx, desc)
		}
		return nil
	}
	postCopy := opts.PostCopy
	opts.PostCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
		logger.Log(ctx, level, "Copied artifact", Descriptor(desc))
		if postCopy != nil {
			return postCopy(ctx, desc)
		}
		return nil
	}
	onMounted := opts.OnMounted
	opts.OnMounted = func(ctx context.Context, desc ocispec.Descriptor) error {
		logger.Log(ctx, level, "Mounted artifact", Descriptor(desc))
		if onMounted != nil {
			return onMounted(ctx, desc)
		}
		return nil
	}
	return *opts
}
