# Builder Notes

## WASM

Idea: Run Homebrew's source using a WASM-compiled Ruby interpreter to build formulae from source.

WASM Ruby:

- [`ruby.wasm`](https://github.com/ruby/ruby.wasm): A first-party WASM build of CRuby
- [How to use Bundler and RubyGems on WebAssembly](https://gist.github.com/kateinoigakukun/5caf3b83b2732b1653e91b0e75ce3390)
- [`ruvy`](https://github.com/Shopify/ruvy): WASM Ruby interpreter by Shopify
- [`webassembly-language-runtimes`](https://github.com/vmware-labs/webassembly-language-runtimes): C of WASM language runtimes
