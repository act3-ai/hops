#!/usr/bin/env bash
# Usage:   ./getinfo.sh <formula> <platform>
# Example: ./getinfo.sh go darwin/arm64

formula=$1
if [[ -z "${formula}" ]]; then
	echo "empty formula argument"
	exit 1
fi

platform=$2
if [[ -z "${platform}" ]]; then
	echo "empty platform argument"
	exit 1
fi

datadir="HOPS_CACHE/data"
mkdir -p "$datadir"

brewinfo="${datadir}/${formula}.brew-info.json"
info="${datadir}/${formula}.curl-info.json"

brew info --json "$formula" >"$brewinfo"
curl -o "$info" "https://formulae.brew.sh/api/formula/$formula.json"

version="$(jq -r '.versions.stable' "$info")"
if [[ -z "${version}" ]]; then
	echo "empty stable version"
	exit 1
fi

revision="$(jq -r '.bottle.stable.rebuild' "$info")"
if [[ "${revision}" != "0" ]]; then
	version="${version}-${revision}"
fi

echo "Fetching ghcr.io/homebrew/core/${formula}:${version}"
index_manifest="${datadir}/${formula}.index.manifest.json"
oras manifest fetch "ghcr.io/homebrew/core/${formula}:${version}" >"$index_manifest"

echo "Fetching ghcr.io/homebrew/core/${formula}:${version} (platform \"${platform}\")"
bottle_manifest="${datadir}/${formula}.${platform//\//_}.manifest.json"
oras manifest fetch --platform="$platform" "ghcr.io/homebrew/core/${formula}:${version}" >"$bottle_manifest"

echo "Fetching sh.brew.tab annotation"
bottle_tab="${datadir}/${formula}.${platform//\//_}.tab.json"
jq '.annotations."sh.brew.tab" | fromjson' "$bottle_manifest" >"$bottle_tab"
