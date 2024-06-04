#!/usr/bin/env bash

dir=${0%/*}

# number of tests to be run
numtests=100

# tool (hops/brew)
# tool="hops"
tool="hops"

# command to benchmark
command="search"
formulae="bash"

# cache setting (warm/cold)
# cache="warm"
cache="cold"

# output JSON file
outfile="${dir}/tests/${command// /_}${formulae:+_}${formulae// /_}.${cache}.${tool}.$(date +%m_%d).json"

# Make sure output file does not already exist
if [ -s "${outfile}" ]; then
	echo "File \"${outfile}\" is not empty"
	exit 1
fi

# Start JSON array
mkdir -p "${dir}/tests"
echo "[" >>"$outfile"

# numtests iterations
for ((i = 1; i <= numtests; i++)); do
	comma=","
	# remove trailing comma after last test result
	[[ $i -eq $numtests ]] && comma=""

	if [[ "${tool}" == "brew" ]]; then
		# Uninstall the formula
		# Uninstall all dependencies
		brew uninstall ${formulae}
		brew autoremove

		# Homebrew cold cache
		if [[ "${cache}" == "cold" ]]; then
			# Remove Homebrew's cache
			# Remove the dummy cache in this directory
			# Homebrew has a bug of some sort that makes it use the env
			# value of "HOMEBREW_CACHE" but only for downloading the API
			rm -rf "$HOME/.cache/Homebrew/"
			rm -rf HOMEBREW_CACHE
		fi
	fi

	if [[ "${tool}" == "hops"* ]]; then
		# Uninstall the formula
		# Uninstall all remaining dependencies
		# hops uninstall ${formulae}
		# hops cleanup
		# rm -rf HOMEBREW_PREFIX # TODO: hops autoremove

		# Hops cold cache
		if [[ "${cache}" == "cold" ]]; then
			# Remove Hops' cache
			rm -rf HOMEBREW_CACHE
			rm -rf HOPS_CACHE
			direnv reload
		fi
	fi

	# shellcheck disable=SC2016
	# time -o "$outfile" -a -f '{"command":"%C","real":%e,"user":%U,"system":%S,"cpu":"%P"},' -- ${tool} images --file internal/benchmark/data/big-brewfile
	command time -o "$outfile" -a -f "{\"date\":\"$(date)\",\"command\":\"%C\",\"real\":%e,\"user\":%U,\"system\":%S,\"cpu\":\"%P\",\"maxrss\":%M}${comma}" -- ${tool} ${command} ${formulae}
	# time -o "$outfile" -a -f "{\"date\":\"$(date)\",\"command\":\"${tool} ${command} --file ${formulae}\",\"real\":%e,\"user\":%U,\"system\":%S,\"cpu\":\"%P\"}${comma}" -- ${tool} ${command} --file "internal/benchmark/data/${formulae}"
	# add trailing comma after previous test result (if there is one)
done

# End JSON array
echo "]" >>"$outfile"
