#!/bin/sh
# Writes the expected output of every test from the problem's reference
# solution: problems/<slug>/tests/NN.out from NN.in through solution.py.
# Pass slugs to redo only those problems; with none, it does them all.
set -eu
cd "$(dirname "$0")/../problems"
[ "$#" -gt 0 ] || set -- *
for slug in "$@"; do
	for in in "$slug"/tests/*.in; do
		python3 "$slug/solution.py" <"$in" >"${in%.in}.out"
	done
	echo "$slug: $(ls "$slug"/tests/*.in | wc -l) tests"
done
