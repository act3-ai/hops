# Notes

## Good Test Case Formulae

- `cowsay`:
  - Current version is 3.04, but it is on "revision 1"
  - Version tags are `3.04_1`
- `brogue`:
  - Large downloads
  - Lots of platform-specific dependencies for Linux
  - Large dependency tree
- `ansible`:
  - Large download (31,172 files, 357.7MB)
- `openssl@3`:
  - Pinned version
  - Registry is `ghcr.io/homebrew/core/openssl/3`
- `libsoxr`:
  - Shows a rebuilt bottle
  - Latest bottle tag is `0.1.3-1`
- `vite`:
  - Depends on node, which is big and includes symlinks in its bottle .tar.gz file
- `gtk+`:
  - Plus in the name is changed to an "x" for the registry path
- `ffmpeg`:
  - Many dependencies
- `bash-completion`:
  - Conflicts with formulae on x86_64 Linux
- `btop`:
  - Might have dependencies reset for linux??
- `clang-format`:
  - Is keg-only on Linux, but not on other system types

## Interesting things

### String sorting

The index returned from formulae.brew.sh is sorted, but not in the same order as the preferred Go sorting

ASCII values:

- `+`: 43
- `-`: 45
- `4`: 52

Homebrew's sorted order of those symbols (incorrect): `-`, `4`, `+`

Go's sorted order of those symbols (correct): `+`, `-`, `4`

Are they treating `+` as `x` in sort order?
