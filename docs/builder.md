# Builder Notes

## Wasm

Idea: Run Homebrew's source using a Wasm-compiled Ruby interpreter to build formulae from source.

Wasm Ruby:

- [`ruby.wasm`](https://github.com/ruby/ruby.wasm): A first-party Wasm build of CRuby
- [How to use Bundler and RubyGems on WebAssembly](https://gist.github.com/kateinoigakukun/5caf3b83b2732b1653e91b0e75ce3390)
- [`ruvy`](https://github.com/Shopify/ruvy): Wasm Ruby interpreter by Shopify
- [`webassembly-language-runtimes`](https://github.com/vmware-labs/webassembly-language-runtimes): C of Wasm language runtimes
