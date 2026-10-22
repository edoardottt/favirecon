#!/bin/bash
#
# https://github.com/edoardottt/favirecon
#
# This script checks if there are duplicate entries (the same hash more than once)
# in the db.json file. JSON parsers (Go's included) silently keep only the last
# value of a duplicate key, so the same hash with two different names is a
# duplicate too, not only two identical lines.
#
# Usage: ./scripts/check-dups.sh [path/to/db.json]
#

file="${1:-pkg/favirecon/db.json}"

if ! command -v jq > /dev/null; then
    echo "[ ERR ] jq NOT FOUND! Install it: https://jqlang.org/download/"
    exit 1
fi

# --stream emits every key as it appears in the file, before jq builds
# the object (which would keep only the last occurrence of each key).
if ! dups=$(jq -nr --stream '
    [inputs | select(length == 2 and (.[0] | length) == 1) | {hash: .[0][0], name: .[1]}]
    | group_by(.hash)[]
    | select(length > 1)
    | "\(.[0].hash): \(map(.name | tostring) | join(" | "))"
' "$file"); then
    echo "[ ERR ] $file IS NOT VALID JSON!"
    exit 1
fi

if [[ -n $dups ]]; then
    echo "[ ERR ] DUPLICATE ENTRIES FOUND!"
    echo "$dups"
    exit 1
else
    echo "[ OK! ] NO DUPLICATES FOUND."
fi
