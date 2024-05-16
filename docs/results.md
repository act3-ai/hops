# Benchmark Results

| Command | Cache | Elapsed | User | System | CPU% | Max RSS | Tests |
| ------- | ----- | ------- | ---- | ------ | ---- | ------- | ----- |
| `brew deps --tree brogue` | warm | 1.06s | 0.75s | 0.13s | 83.34% | no data | 100 |
| `hops deps --tree brogue` | warm | 0.27s | 0.19s | 0.01s | 74.43% | no data | 100 |
| `brew deps --tree ffmpeg` | cold | 1.83s | 0.83s | 0.22s | 57.14% | no data | 29 |
| `hops deps --tree ffmpeg` | cold | 0.66s | 0.28s | 0.03s | 49.33% | no data | 27 |
| `brew deps --tree ffmpeg` | warm | 0.99s | 0.72s | 0.13s | 86.07% | no data | 30 |
| `hops deps --tree ffmpeg` | warm | 0.23s | 0.19s | 0s | 84.79% | no data | 29 |
| `brew deps brogue` | cold | 2.03s | 0.9s | 0.22s | 54.97% | no data | 29 |
| `hops deps brogue` | cold | 0.67s | 0.29s | 0.04s | 49.75% | no data | 24 |
| `brew deps brogue` | warm | 1.13s | 0.8s | 0.14s | 83.46% | no data | 100 |
| `hops deps brogue` | warm | 0.27s | 0.19s | 0.01s | 74.47% | no data | 99 |
| `brew deps ffmpeg` | cold | 1.97s | 0.87s | 0.21s | 55.32% | no data | 28 |
| `hops deps ffmpeg` | cold | 0.58s | 0.29s | 0.03s | 56.88% | no data | 51 |
| `hops deps ffmpeg` | cold | 0.64s | 0.27s | 0.03s | 48.12% | no data | 26 |
| `brew deps ffmpeg` | warm | 1.03s | 0.76s | 0.13s | 85.97% | no data | 30 |
| `hops deps ffmpeg` | warm | 0.24s | 0.19s | 0.01s | 84.73% | no data | 60 |
| `hops deps ffmpeg` | warm | 0.23s | 0.19s | 0s | 84.43% | no data | 30 |
| `hops images --file asce-tools` | cold | 3.22s | 0.32s | 0.07s | 11.85% | no data | 26 |
| `hops images --file asce-tools` | cold | 2.07s | 0.35s | 0.07s | 20.17% | no data | 29 |
| `hops images --file asce-tools` | warm | 1.73s | 0.24s | 0.04s | 16.19% | no data | 31 |
| `hops images --file asce-tools` | warm | 2.6s | 0.23s | 0.03s | 10% | no data | 28 |
| `hops images --file big-brewfile` | cold | 8.58s | 0.54s | 0.18s | 7.92% | no data | 38 |
| `hops images --file big-brewfile` | warm | 8.02s | 0.41s | 0.16s | 6.96% | no data | 28 |
| `hops images --file big-brewfile` | warm | 8.19s | 0.41s | 0.16s | 6.87% | no data | 31 |
| `hops images --file big-brewfile` | warm | 26.09s | 0.38s | 0.15s | 1.9% | no data | 10 |
| `hops images poetry` | cold | 1.91s | 0.27s | 0.03s | 16.93% | no data | 27 |
| `hops images poetry` | warm | 0.99s | 0.2s | 0.01s | 21.72% | no data | 96 |
| `brew install brogue` | cold | 33.25s | 5.21s | 6.12s | 33.56% | no data | 27 |
| `hops install brogue` | cold | 0.81s | 0.68s | 0.32s | 124.89% | no data | 38 |
| `hops install brogue` | cold | 0.79s | 0.68s | 0.32s | 126.11% | no data | 38 |
| `hops install brogue` | cold | 1.91s | 0.65s | 0.31s | 50.46% | no data | 28 |
| `brew install brogue` | warm | 13.15s | 4.69s | 5.38s | 76.24% | no data | 29 |
| `hops install brogue` | warm | 0.87s | 0.64s | 0.45s | 126.64% | no data | 22 |
| `hops--link-conc install ffmpeg` | cold | 6.12s | 8.52s | 5.45s | 228.18% | no data | 39 |
| `hops install ffmpeg` | cold | 5.97s | 8.19s | 5.35s | 226.43% | no data | 37 |
| `brew install gh` | cold | 3.84s | 1.24s | 0.48s | 44.89% | no data | 88 |
| `hops install gh` | cold | 1.51s | 0.43s | 0.09s | 34.26% | no data | 91 |
| `brew install gh` | warm | 1.15s | 0.8s | 0.15s | 82.41% | no data | 100 |
| `hops install gh` | warm | 0.92s | 0.42s | 0.1s | 56.92% | no data | 84 |
| `brew search ansible` | cold | 2.34s | 0.98s | 0.24s | 52.8% | no data | 30 |
| `hops search ansible` | cold | 0.68s | 0.27s | 0.03s | 45.15% | no data | 26 |
| `brew search ansible` | warm | 1.15s | 0.82s | 0.14s | 84.19% | no data | 100 |
| `hops search ansible` | warm | 0.24s | 0.19s | 0.01s | 84.91% | no data | 57 |
| `hops search ansible` | warm | 0.26s | 0.19s | 0s | 75.31% | no data | 100 |
| `hops--no-cached-names search git` | warm | 0.24s | 0.19s | 0.01s | 85.63% | no data | 30 |
| `brew search zig` | cold | 2.28s | 0.93s | 0.23s | 51.72% | no data | 29 |
| `brew search zig` | warm | 1.14s | 0.81s | 0.14s | 83.37% | no data | 100 |
| `hops--cached-names search zig` | warm | 0.23s | 0.18s | 0.01s | 85.1% | no data | 30 |
| `hops--no-cached-names search zig` | warm | 0.22s | 0.18s | 0s | 87.03% | no data | 30 |
| `hops search zig` | warm | 0.24s | 0.19s | 0.01s | 84.71% | no data | 58 |
| `hops search zig` | warm | 0.26s | 0.19s | 0s | 75.08% | no data | 98 |
| `hops xinstall brogue` | cold | 2.37s | 0.66s | 0.53s | 49.59% | no data | 32 |
| `hops xinstall brogue` | cold | 2.27s | 0.67s | 0.54s | 53.15% | no data | 39 |
| `hops xinstall brogue` | warm | 1.24s | 0.26s | 0.21s | 38.03% | no data | 39 |
| `hops xinstall brogue` | warm | 1.26s | 0.65s | 0.49s | 90.5% | no data | 34 |
| `hops xinstall gh` | cold | 1.29s | 0.26s | 0.12s | 29.48% | no data | 40 |
| `hops xinstall gh` | cold | 1.37s | 0.26s | 0.12s | 27.86% | 24668kb | 36 |
| `hops xinstall gh` | cold | 1.55s | 0.25s | 0.09s | 21.87% | no data | 30 |
| `hops xinstall gh` | cold | 2.23s | 0.25s | 0.09s | 21.26% | no data | 31 |
| `hops xinstall gh` | cold | 1.63s | 0.23s | 0.08s | 19.04% | no data | 26 |
| `hops xinstall gh` | warm | 0.67s | 0.22s | 0.06s | 43.1% | no data | 40 |
| `hops xinstall gh` | warm | 0.62s | 0.21s | 0.06s | 45% | 22686kb | 30 |
| `hops xinstall gh` | warm | 0.73s | 0.21s | 0.07s | 38.13% | no data | 30 |
| `hops xinstall gh` | warm | 0.76s | 0.22s | 0.07s | 38.9% | no data | 31 |
| `hops xinstall gh` | warm | 0.97s | 0.21s | 0.07s | 28.9% | no data | 29 |
