#!/bin/bash
echo -e "23aafe84f813bd5093599691ea5731425effe1b8c3f7c1e3c049012558160b8c\n617af9da0ea3053e89ce43e1514640bcd41a3bd69810ca0f0d82e5cf8b4cadb8" > ../bin/genesis_hashes.txt && \
go run ../cmd/settings/main.go --port 35000 --interval 10s --localhost=false && \
./build.sh && \
docker-compose -f ../docker-compose.ci.yml build && \
docker-compose -f ../docker-compose.ci.yml up -d