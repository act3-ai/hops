#!/usr/bin/env bash
# Usage:   ./getallinfo.sh

datadir="HOPS_CACHE/data"
mkdir -p "$datadir"

info="${datadir}/formula.json"

curl -o "$info" "https://formulae.brew.sh/api/formula.json"
