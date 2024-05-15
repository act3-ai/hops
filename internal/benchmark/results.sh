#!/usr/bin/env bash

outfile="internal/benchmark/results.md"

cat >"${outfile}" <<EOF
# Benchmark Results

| Command | Cache | Elapsed | User | System | CPU% | Max RSS | Tests |
| ------- | ----- | ------- | ---- | ------ | ---- | ------- | ----- |
EOF

for file in internal/benchmark/tests/*.json; do
	# echo "${file}:"

	cache="warm"
	[[ "${file}" == *.cold.*.json ]] && cache="cold"

	cmd=$(jq -r '.[0].command' "$file")
	tests=$(jq 'length' "$file")
	real=$(jq 'length as $l | reduce .[] as $item (0; . + $item.real) / $l | .*100|round/100' "$file")
	user=$(jq 'length as $l | reduce .[] as $item (0; . + $item.user) / $l | .*100|round/100' "$file")
	system=$(jq 'length as $l | reduce .[] as $item (0; . + $item.system) / $l | .*100|round/100' "$file")
	cpu=$(jq 'length as $l | reduce .[] as $item (0; . + ($item.cpu|rtrimstr("%")|tonumber)) / $l | .*100|round/100' "$file")
	maxrss=$(jq 'length as $l | reduce .[] as $item (0; . + $item.maxrss) / $l | .|round' "$file")
	if (( maxrss == 0 )); then
		maxrss="no data"
	else
		maxrss="${maxrss}kb"
	fi

	echo "${cmd}: cache(${cache}) real(${real}) user(${user}) system(${system}) maxrss(${maxrss})"

	echo "| \`${cmd}\` | ${cache} | ${real}s | ${user}s | ${system}s | ${cpu}% | ${maxrss} | ${tests} |" >>"${outfile}"
done
