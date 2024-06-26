---
title: hops completion fish
description: Generate the autocompletion script for fish
---

<!--
This documentation is auto generated by a script.
Please do not edit this file directly.
-->

<!-- markdownlint-disable-next-line single-title -->
# hops completion fish

Generate the autocompletion script for fish

## Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	hops completion fish | source

To load completions for every new session, execute once:

	hops completion fish > ~/.config/fish/completions/hops.fish

You will need to start a new shell for this setup to take effect.


## Usage

```plaintext
hops completion fish [flags]
```

## Options

```plaintext
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

## Options inherited from parent commands

```plaintext
      --concurrency int   Concurrency level (default 8)
      --config strings    Set config file search paths (default `hops-config.yaml`,`$XDG_CONFIG_HOME/hops/config.yaml`,`/etc/hops/config.yaml`)
  -d, --debug count       Display more debugging information
      --log-fmt string    Set format for log messages. Options: text, json (default "text")
  -q, --quiet count       Make some output more quiet
  -v, --verbose count     Make some output more verbose
```
