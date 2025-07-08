#!/bin/sh
# fix_whitespace.sh - Find files and fix trailing whitespace and EOF newlines
#
# Usage: fix_whitespace.sh [find arguments]
#        fix_whitespace.sh -- file1 file2 ...
#
# Automatically prunes .git and node_modules directories
#
# Examples:
#   fix_whitespace.sh . -name '*.go' -o -name '*.md'
#   fix_whitespace.sh src/ -name '*.js'
#   fix_whitespace.sh -- README.md LICENCE.txt

set -eu

# Function to fix a single file
fix_file() {
	local file="$1" last_byte

	# Skip if not a regular file
	[ -f "$file" ] || return 0

	# Remove trailing whitespace
	sed -i 's/[[:space:]]*$//' "$file"

	# Leave empty files alone
	[ -s "$file" ] || return 0


	# Check last byte to see if file ends with newline
	# Use od to get hexadecimal representation of last byte
	last_byte=$(tail -c 1 "$file" | od -An -tx1 | tr -d ' \t')

	# If last byte is not newline (0x0a), add one
	if [ "x$last_byte" != "x0a" ]; then
		printf '\n' >> "$file"
	elif [ "$(wc -c < "$file")" -eq 1 ]; then
		# File only contains a newline, truncate it
		: > "$file"
	fi
}

# Check for explicit file mode (starts with --)
if [ $# -gt 0 ] && [ "$1" = "--" ]; then
	# Explicit file mode
	shift
	for file in "$@"; do
		fix_file "$file"
	done
else
	# Find mode
	# Default to current directory if no args
	if [ $# -eq 0 ]; then
		set -- .
	fi

	# Execute find with forced -type f and auto-pruning
	# Pass arguments directly to find, preserving quotes
	find "$@" \( -name .git -o -name node_modules \) -prune -o -type f -print | while IFS= read -r file; do
		fix_file "$file"
	done
fi
