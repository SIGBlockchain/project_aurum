#!/bin/bash
ROOT=$1
if [[ -z "$ROOT" ]]; then 
    echo "ERROR: Root directory as argument required"
    exit 1
fi
echo -e "23aafe84f813bd5093599691ea5731425effe1b8c3f7c1e3c049012558160b8c\n617af9da0ea3053e89ce43e1514640bcd41a3bd69810ca0f0d82e5cf8b4cadb8" > bin/genesis_hashes.txt
go run $1/cmd/settings/main.go --port 35000 --interval 10s --localhost=false