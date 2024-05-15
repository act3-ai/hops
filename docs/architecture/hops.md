# Hops Design

![hops flower](https://media.istockphoto.com/id/1200973259/vector/vector-illustration-green-cone-of-hop-symbol-of-beer-pub-and-alcoholic-beverage-graphic.jpg?s=612x612&w=0&k=20&c=JDRjkdkf5C9BxWHO3RINDFjpKD9a-1-NbO_dFWH6Uzc=)

---

## Hops Goals

The goal of the Hops project is to create a lightweight client for installing, upgrading, and mirroring Homebrew bottles

Goals:

- OCI native
- Follow 12-factor app best practices
- Full compatibility with `ghcr.io/Homebrew/core` registry
- Share Homebrew's local cache
  - Hops will use files cached by Homebrew and vice versa
- Lightweight and performant
- Work independent of Homebrew

---

## Plans

- Minimum functionality for Homebrew independence
  - Install, upgrade, uninstall, search, shell hooks, configuration
- Implement mirroring features
  - List images to mirror
  - Support installing from mirrored Homebrew bottle registry
- Stretch: experiment with Homebrew integrations

---

## Implementation Details

Hops will be written in Go for the following reasons:

- I am already familiar with it
- It has language-level and standard library support for concurrency
- It is a compiled language with better performance than Ruby
<!-- - Only systems languages (such as C++ and Rust) offer consistently better performance -->
- Supports many platforms

The implementation will rely on the following Go packages:

- [ORAS](https://github.com/oras-project/oras-go): OCI artifact transfers
- [sourcegraph/conc](https://github.com/sourcegraph/conc): structured concurrency

---

## Early Benchmarks

Below are early benchmarks comparing equivalent Hops and Homebrew commands. The initial implementations of these commands in Hops will continue to be optimized over time.

> Benchmarks were run on a MacBook Air M1. Hops' concurrency level was set at 8 for all tests.

---

## Benchmark 1

```mermaid
---
config:
  xyChart:
    xAxis:
      showTick: false
  themeVariables:
    xyChart:
      plotColorPalette: "#2e2a24, #be862d, #2e2a24, #be862d"
---
xychart-beta
    title "Dependency Resolution: ffmpeg (105 dependencies)"
    x-axis ["brew", "hops", "brew [cache]", "hops [cache]"]
    y-axis "seconds" 0 --> 2.5
    bar [1.97, 0.64, 1.03, .23]
```

---

## Benchmark 2

```mermaid
---
config:
  xyChart:
    xAxis:
      showTick: false
  themeVariables:
    xyChart:
      plotColorPalette: "#2e2a24, #be862d, #2e2a24, #be862d"
---
xychart-beta
    title "Search: ansible"
    x-axis ["brew", "hops", "brew [cache]", "hops [cache]"]
    y-axis "seconds" 0 --> 2.5
    bar [2.34, 0.68, 1.15, 0.26]
```

---

## Benchmark 3

```mermaid
---
config:
  xyChart:
    xAxis:
      showTick: false
  themeVariables:
    xyChart:
      plotColorPalette: "#2e2a24, #be862d, #2e2a24, #be862d"
---
xychart-beta
    title "Install: gh (no dependencies)"
    x-axis ["brew", "hops", "brew [cache]", "hops [cache]"]
    y-axis "seconds" 0 --> 4
    bar [3.84, 1.51, 1.15, 0.92]
```

---

## Benchmark 4

```mermaid
---
config:
  xyChart:
    xAxis:
      showTick: false
  themeVariables:
    xyChart:
      plotColorPalette: "#2e2a24, #be862d, #2e2a24, #be862d"
---
xychart-beta
    title "Install: brogue (19 dependencies)"
    x-axis ["brew", "hops", "brew [cache]", "hops [cache]"]
    y-axis "seconds" 0 --> 35
    bar [33.25, 1.91, 13.15, 0.87]
```
