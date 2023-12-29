#!/bin/bash

numerals=("first" "second" "third")

for dir in ./*; do
	if [ -d "$dir" ]; then
		files=("$dir"/*)
		#printf "{\"contents of %s\": \"%s\", \n" $dir $files
		file_count="$((${#files[@]}))"
		#printf "\"count\": \"%d\", \n" $file_count
		for ((i = 0; i < file_count; i++)); do
			f="${files[$i]}"
			#printf "\"working with\": \"%s\",\n" $f
			base="$(basename "$f")"
			#printf "\"base name\": \"%s\",\n" $base
			file_head="$(echo "$base" | awk -F '.' ' { print $1 }')"
			file_tail="$(echo "$base" | cut -d"." -f2-)"
			in=$((i / 2 + 1))
			idx=$((in - 1))
			#printf "\"index (incremented by 1)\": \"%d\"},\n" $idx
			#	printf "would rename %s to %s/%d_%s_migration.%s" "$f" "$dir" $idx "${numerals[$i]}" "$file_tail"
			git mv "$f" "$dir"/${in}_"${numerals[$idx]}"_migration.$file_tail
		done
	fi
done
