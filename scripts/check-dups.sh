#!/bin/bash
#
# https://github.com/edoardottt/favirecon
#
# This script checks if there are duplicate entries in the db.json file.
#

file="pkg/favirecon/db.json"

dups=$(cat $file | sort | uniq -d)

if [[ -n $dups ]]; then
    echo "[ ERR ] DUPLICATE ENTRIES FOUND!"
    echo "$dups"
    exit 1
else
    echo "[ OK! ] NO DUPLICATES FOUND."
fi