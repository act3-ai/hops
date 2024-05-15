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

func OCIPlatformValue(plat *ocispec.Platform) slog.Attr {
	if plat == nil {
		return slog.String("platform", "nil")
	}
	return slog.Attr{
		Key:   "platform",
		Value: slog.GroupValue(ociPlatformAttrs(*plat)...),
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

func Descriptor(desc ocispec.Descriptor) slog.Attr {
	return slog.Attr{
		Key:   "desc",
		Value: slog.GroupValue(descriptorValues(desc)...),
	}
}

// // Custom log levels for higher-level debug prints
// // Double the normal debug level
// Debug2 = log.DebugLevel * 2
// // Triple the normal debug level
// Debug3 = log.DebugLevel * 3

func WithLogging(logger *slog.Logger, level slog.Level, opts *oras.CopyGraphOptions) oras.CopyGraphOptions {
	opts.OnCopySkipped = func(ctx context.Context, desc ocispec.Descriptor) error {
		logger.Log(ctx, level, "Skipped artifact", Descriptor(desc))
		return nil
	}
	opts.PostCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
		logger.Log(ctx, level, "Copied artifact", Descriptor(desc))
		return nil
	}
	opts.OnMounted = func(ctx context.Context, desc ocispec.Descriptor) error {
		logger.Log(ctx, level, "Mounted artifact", Descriptor(desc))
		return nil
	}
	return *opts
}
