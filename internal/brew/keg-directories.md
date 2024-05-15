# Keg Directories

**name**: function name in the Homebrew Ruby API

| name              | default path                                   | example                                                     |
| ----------------- | ---------------------------------------------- | ----------------------------------------------------------- |
| `HOMEBREW_PREFIX` | output of `$(brew --prefix)`                   | `/usr/local`                                                |
| `prefix`          | `#{HOMEBREW_PREFIX}/Cellar/#{name}/#{version}` | `/usr/local/Cellar/foo/0.1`                                 |
| `opt_prefix`      | `#{HOMEBREW_PREFIX}/opt/#{name}`               | `/usr/local/opt/foo`                                        |
| `bin`             | `#{prefix}/bin`                                | `/usr/local/Cellar/foo/0.1/bin`                             |
| `doc`             | `#{prefix}/share/doc/#{name}`                  | `/usr/local/Cellar/foo/0.1/share/doc/foo`                   |
| `include`         | `#{prefix}/include`                            | `/usr/local/Cellar/foo/0.1/include`                         |
| `info`            | `#{prefix}/share/info`                         | `/usr/local/Cellar/foo/0.1/share/info`                      |
| `lib`             | `#{prefix}/lib`                                | `/usr/local/Cellar/foo/0.1/lib`                             |
| `libexec`         | `#{prefix}/libexec`                            | `/usr/local/Cellar/foo/0.1/libexec`                         |
| `man`             | `#{prefix}/share/man`                          | `/usr/local/Cellar/foo/0.1/share/man`                       |
| `man`             | `[1-8] #{prefix}/share/man/man[1-8]`           | `/usr/local/Cellar/foo/0.1/share/man/man[1-8]`              |
| `sbin`            | `#{prefix}/sbin`                               | `/usr/local/Cellar/foo/0.1/sbin`                            |
| `share`           | `#{prefix}/share`                              | `/usr/local/Cellar/foo/0.1/share`                           |
| `pkgshare`        | `#{prefix}/share/#{name}`                      | `/usr/local/Cellar/foo/0.1/share/foo`                       |
| `elisp`           | `#{prefix}/share/emacs/site-lisp/#{name}`      | `/usr/local/Cellar/foo/0.1/share/emacs/site-lisp/foo`       |
| `frameworks`      | `#{prefix}/Frameworks`                         | `/usr/local/Cellar/foo/0.1/Frameworks`                      |
| `kext_prefix`     | `#{prefix}/Library/Extensions`                 | `/usr/local/Cellar/foo/0.1/Library/Extensions`              |
| `zsh_function`    | `#{prefix}/share/zsh/site-functions`           | `/usr/local/Cellar/foo/0.1/share/zsh/site-functions`        |
| `fish_function`   | `#{prefix}/share/fish/vendor_functions`        | `/usr/local/Cellar/foo/0.1/share/fish/vendor_functions`     |
| `bash_completion` | `#{prefix}/etc/bash_completion.d`              | `/usr/local/Cellar/foo/0.1/etc/bash_completion.d`           |
| `zsh_completion`  | `#{prefix}/share/zsh/site-functions`           | `/usr/local/Cellar/foo/0.1/share/zsh/site-functions`        |
| `fish_completion` | `#{prefix}/share/fish/vendor_completions.d`    | `/usr/local/Cellar/foo/0.1/share/fish/vendor_completions.d` |
| `etc`             | `#{HOMEBREW_PREFIX}/etc`                       | `/usr/local/etc`                                            |
| `pkgetc`          | `#{HOMEBREW_PREFIX}/etc/#{name}`               | `/usr/local/etc/foo`                                        |
| `var`             | `#{HOMEBREW_PREFIX}/var`                       | `/usr/local/var`                                            |
| `buildpath`       | temporary directory somewhere on your system   | `/private/tmp/[formula-name]-0q2b/[formula-name]`           |

## Info Mode

- `info/([^.].*?\.info|dir)$`

## Explicit Directories

- etc/*
- share/
- share/locale/`([a-z]{2}|C|POSIX)(_[A-Z]{2})?(\.[a-zA-Z\-0-9]+(@.+)?)?`
- share/man/`([a-z]{2}|C|POSIX)(_[A-Z]{2})?(\.[a-zA-Z\-0-9]+(@.+)?)?`
- share/icons
- share/zsh
- share/fish
- share/lua
- share/guile
- share/(any of `KegSharePaths`)
- lib/
- lib/pkgconfig
- lib/cmake
- lib/dtrace
- lib/gdk-pixbuf
- lib/ghc
- lib/gio
- lib/lua
- lib/mecab
- lib/node
- lib/ocaml
- lib/perl5
- lib/php
- lib/python2.d
- lib/python3.d
- lib/R
- lib/ruby
- Frameworks
- Frameworks/*.framework
- Frameworks/*.framework/Versions

## Skipped Directories

- bin/*
- sbin/*

## Skipped Files

- share/locale/locale.alias
- share/icons/*/icon-theme.cache
- lib/charset.alias

## Linked Files

- bin/* (files)
- sbin/* (files)
- include/*
- share/* (else)
- lib/* (else)
- Frameworks/* (else)
