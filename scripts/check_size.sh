#!/bin/bash
set -e
files=$(find . -name '*.go' ! -name '*_test.go' ! -path './vendor/*' | sort)
count=$(printf '%s\n' "$files" | sed '/^$/d' | wc -l | tr -d ' ')
lines=$(wc -l $files | tail -1 | awk '{print $1}')
echo "Go files=$count lines=$lines"
if [ "$count" -le 20 ] || [ "$count" -ge 25 ]; then echo "warning: documented target is 21-24 files"; fi
if [ "$lines" -le 2000 ] || [ "$lines" -ge 2200 ]; then echo "warning: documented target is 2001-2199 lines"; fi
