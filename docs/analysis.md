# Benchmark Analysis

For all benchmarks, tests with outlier CPU percentages were removed before averaging.

> Hops' performance can improve greatly with improved support for Homebrew's existing data surfaces. Currently Hops downloads and caches the entire formulae.brew.sh formula index, which is around 25MB of JSON. Once Homebrew's "bottle tabs" (an image manifest annotation) are supported by Hops, only 100kB of data will be downloaded or loaded from the cache.

## Package Size Comparison

| Tool            | # Files    | Size  |
| --------------- | ---------- | ----- |
| `brew`          | 2100 files | 110MB |
| `hops`          | 1 file     | 14MB  |
| April 9: `hops` | 1 file     | 10MB  |

Hops is a 10MB statically compiled binary.

Homebrew is distributed as a 110MB Git repository. On some systems, Homebrew will also install a version of Ruby. In all, the `HOMEBREW_PREFIX/Library/` directory containing the majority of the Ruby source code is 87MB, not including any Homebrew Taps that are cloned during use.

## Dependency Resolving

Hops is faster at resolving dependencies by a factor of 3-4x.

brogue (19 dependencies):

| Command                   | Cache | Elapsed | User  | System | CPU%   | Tests |
| ------------------------- | ----- | ------- | ----- | ------ | ------ | ----- |
| `brew deps brogue`        | cold  | 2.03s   | 0.9s  | 0.22s  | 54.97% | 29    |
| `hops deps brogue`        | cold  | 0.69s   | 0.29s | 0.04s  | 49%    | 26    |
| `brew deps brogue`        | warm  | 1.13s   | 0.8s  | 0.14s  | 83.46% | 100   |
| `hops deps brogue`        | warm  | 0.27s   | 0.19s | 0.01s  | 74.31% | 100   |
| `brew deps --tree brogue` | warm  | 1.06s   | 0.75s | 0.13s  | 83.34% | 100   |
| `hops deps --tree brogue` | warm  | 0.27s   | 0.19s | 0.01s  | 74.43% | 100   |

ffmpeg (105 dependencies):

| Command                 | Cache | Elapsed | User  | System | CPU%   | Tests |
| ----------------------- | ----- | ------- | ----- | ------ | ------ | ----- |
| `brew deps ffmpeg`      | cold  | 2.22s   | 0.87s | 0.21s  | 53.03% | 30    |
| `hops deps ffmpeg`      | cold  | 0.74s   | 0.27s | 0.03s  | 45.17% | 30    |
| 4/9: `hops deps ffmpeg` | cold  | 0.58s   | 0.29s | 0.03s  | 56.88% | 51    |
| `brew deps ffmpeg`      | warm  | 1.03s   | 0.76s | 0.13s  | 85.97% | 30    |
| `hops deps ffmpeg`      | warm  | 0.23s   | 0.19s | 0s     | 84.43% | 30    |
| 4/9: `hops deps ffmpeg` | warm  | 0.24s   | 0.19s | 0.01s  | 84.73% | 60    |

## Searching

Checked both A and Z just to make sure. Factor of 4 again.

| Command                    | Cache | Elapsed | User  | System | CPU%   | Tests |
| -------------------------- | ----- | ------- | ----- | ------ | ------ | ----- |
| `brew search ansible`      | warm  | 1.15s   | 0.82s | 0.14s  | 84.19% | 100   |
| `hops search ansible`      | warm  | 0.26s   | 0.19s | 0s     | 75.31% | 100   |
| 4/9: `hops search ansible` | warm  | 0.24s   | 0.19s | 0.01s  | 84.91% | 57    |
| `brew search zig`          | warm  | 1.14s   | 0.81s | 0.14s  | 83.37% | 100   |
| `hops search zig`          | warm  | 0.26s   | 0.19s | 0s     | 74.87% | 100   |
| 4/9: `hops search zig`     | warm  | 0.24s   | 0.19s | 0.01s  | 84.71% | 58    |

## Installation

Formula with no dependencies:

Factor of 4 for cold cache, factor of :shrug: for warm cache. Since Hops has the benefit of the Go runtime, I'm going to call its warm cache performance worse.

| Command           | Cache | Elapsed | User  | System | CPU%   | Memory   | Tests |
| ----------------- | ----- | ------- | ----- | ------ | ------ | -------- | ----- |
| `brew install gh` | cold  | 4.16s   | 1.24s | 0.48s  | 42.8%  |          | 100   |
| `hops install gh` | cold  | 1.62s   | 0.43s | 0.09s  | 33.17% |          | 100   |
| `hops install gh` | cold  | 1.85s   | 0.53s | 0.12s  | 36.33% | 110896kb | 36    |
| `brew install gh` | warm  | 1.15s   | 0.8s  | 0.15s  | 82.41% |          | 100   |
| `hops install gh` | warm  | 1.02s   | 0.42s | 0.1s   | 53.66% |          | 100   |

Formula with many dependencies:

Factor of 16 for cold cache (concurrency!), factor of 13 for warm cache (concurrency again?).

| Command                      | Cache    | Elapsed | User  | System | CPU%    | Tests |
| ---------------------------- | -------- | ------- | ----- | ------ | ------- | ----- |
| `brew install brogue`        | cold     | 33.25s  | 5.21s | 6.12s  | 33.56%  | 27    |
| `hops install brogue`        | cold     | 2.02s   | 0.65s | 0.31s  | 48.9%   | 30    |
| `brew install brogue`        | warm     | 13.15s  | 4.69s | 5.38s  | 76.24%  | 29    |
| `hops install brogue`        | warm     | 1.11s   | 0.65s | 0.45s  | 113.8%  | 30    |
| 4/9: `hops install brogue`   | warm-ish | 0.79s   | 0.68s | 0.32s  | 126.11% | 38    |
| 4/11: `hops xinstall brogue` | cold     | 2.37s   | 0.66s | 0.53s  | 49.59%  | 32    |
| 4/11: `hops xinstall brogue` | warm     | 1.24s   | 0.26s | 0.21s  | 38.03%  | 39    |
| 4/12: `hops xinstall brogue` | cold     | 2.27s   | 0.67s | 0.54s  | 53.15%  | 39    |
| 4/12: `hops xinstall brogue` | warm     | 1.26s   | 0.65s | 0.49s  | 90.5%   | 34    |

comparing linking with concurrency:

| Command                           | Cache    | Elapsed | User  | System | CPU%    | Tests |
| --------------------------------- | -------- | ------- | ----- | ------ | ------- | ----- |
| concurrent: `hops install ffmpeg` | warm-ish | 6.12s   | 8.52s | 5.45s  | 228.18% | 39    |
| not: `hops install ffmpeg`        | warm-ish | 5.97s   | 8.19s | 5.35s  | 226.43% | 37    |

> CPU percentage of 113 for the warm cache install with brogue. It looks like network access drops the process to the background, but reading files from the cache means we are free to hog CPU.

## Alternate Install Methods

Hops' `xinstall` command is being used for experimenting for other bottle installation methods.

### Manifest-first method

This method starts by downloading the manifest of the latest tag in a repository, and uses the `sh.brew.tab` annotation alone.

By listing all tags and identifying the latest tags:

| Date | Command            | Cache | Elapsed | CPU%   | Memory  | Tests |
| ---- | ------------------ | ----- | ------- | ------ | ------- | ----- |
|      | `hops xinstall gh` | cold  | 1.55s   | 21.87% |         | 30    |
|      | `hops xinstall gh` | warm  | 0.73s   | 38.13% |         | 30    |
| 4/11 | `hops xinstall gh` | cold  | 1.29s   | 29.48% |         | 40    |
| 4/11 | `hops xinstall gh` | warm  | 0.67s   | 43.1%  |         | 40    |
| 4/12 | `hops xinstall gh` | cold  | 1.37s   | 27.86% | 24668kb | 36    |
| 4/12 | `hops xinstall gh` | warm  | 0.62s   | 45%    | 22686kb | 30    |

This test surprised me, I thought it would be quicker for the cold cache. I tried it again but instead with the tag hardcoded, here are those results:

| Command            | Cache | Elapsed | User  | System | CPU%   | Tests |
| ------------------ | ----- | ------- | ----- | ------ | ------ | ----- |
| `hops xinstall gh` | cold  | 1.63s   | 0.23s | 0.08s  | 19.04% | 26    |
| `hops xinstall gh` | warm  | 0.97s   | 0.21s | 0.07s  | 28.9%  | 29    |

Without setting GOMAXPROCS:

| Command            | Cache | Elapsed | User  | System | CPU%   | Tests |
| ------------------ | ----- | ------- | ----- | ------ | ------ | ----- |
| `hops xinstall gh` | cold  | 2.23s   | 0.25s | 0.09s  | 21.26% | 31    |
| `hops xinstall gh` | warm  | 0.76s   | 0.22s | 0.07s  | 38.9%  | 31    |

## Listing Images

[big-brewfile](../test/benchmark/data/big-brewfile):

| Command                                   | Cache | Elapsed | User  | System | CPU%  | Tests |
| ----------------------------------------- | ----- | ------- | ----- | ------ | ----- | ----- |
| `hops images --file big-brewfile`         | warm  | 26.09s  | 0.38s | 0.15s  | 1.9%  | 10    |
| `hops images --file big-brewfile`         | warm  | 8.19s   | 0.41s | 0.16s  | 6.87% | 31    |
| New: `hops images --file big-brewfile`    | warm  | 8.02s   | 0.41s | 0.16s  | 6.96% | 28    |
| 4/9/24: `hops images --file big-brewfile` | cold  | 8.58s   | 0.54s | 0.18s  | 7.92% | 38    |

[asce-tools](../test/benchmark/data/asce-tools):

| Command                                | Cache | Elapsed | User  | System | CPU%   | Tests |
| -------------------------------------- | ----- | ------- | ----- | ------ | ------ | ----- |
| `hops images --file asce-tools`        | cold  | 3.22s   | 0.32s | 0.07s  | 11.85% | 26    |
| New: `hops images --file asce-tools`   | cold  | 2.07s   | 0.35s | 0.07s  | 20.17% | 29    |
| `hops images --file asce-tools`        | warm  | 2.78s   | 0.23s | 0.03s  | 9.67%  | 30    |
| New:  `hops images --file asce-tools`  | warm  | 1.73s   | 0.24s | 0.04s  | 16.19% | 31    |
