#!/bin/bash

numerals=("first" "second" "third")

for dir in ./*; do
	if [ -d "$dir" ]; then
		fc="$(fd -t f . "$dir" 2>/dev/null | rg -v "total " | wc -l)"
		file_count="$(($fc / 2))"
		for f in "$dir"/*; do
			for ((i = 0; i < file_count; i++)); do
				base="$(basename "$f")"
				file_head="$(echo "$base" | awk -F '.' ' { print $1 }')"
				file_tail="$(echo "$base" | cut -d"." -f2-)"
				idx=$((i + 1))
				#	printf "would rename %s to %s/%d_%s_migration.%s" "$f" "$dir" $idx "${numerals[$i]}" "$file_tail"
				git mv -vn "$f" "$dir"/${idx}_"${numerals[$i]}"_migration.$file_tail
			done
		done
	fi
done
