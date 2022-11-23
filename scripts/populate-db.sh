#!/bin/bash


# check duplicates:
nHash=$(cat hash-db.csv | cut -d "," -f 1 | wc -l)
nName=$(cat hash-db.csv | cut -d "," -f 2 | wc -l)

nUniqHash=$(cat hash-db.csv | cut -d "," -f 1 | sort -u | wc -l)
nUniqName=$(cat hash-db.csv | cut -d "," -f 2 | sort -u | wc -l)

if [[ "$nHash" != "$nUniqHash" ]]; then
    echo "[ERROR] Duplicated hashes!"
    exit 1
fi

if [[ "$nName" != "$nUniqName" ]]; then
    echo "[WARNING] Duplicated names!"
fi


touch db.go

echo "package favirecon

var (
	db = map[string]string{" > db.go

while IFS= read -r line
do
  echo "$line" 
        hash=$(echo "${line}" | cut -f 1 -d ,)
        name=$(echo "${line}" | cut --complement -f 1 -d ,)
        echo -n '		"' >> db.go
        echo -n "$hash" >> db.go
        echo -n '": "' >> db.go
        echo -n "$name" >> db.go
        echo '",' >> db.go
done < "hash-db.csv"

echo "	}
)
" >> db.go