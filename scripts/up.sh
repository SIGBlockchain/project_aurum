#!/bin/bash
echo "Changing port to 26000, interval to 12 hours, and localhost to false"
go run ../cmd/settings/main.go --port 26000 --interval 12h --localhost=false
docker-compose -f ../docker-compose.yml build
docker-compose -f ../docker-compose.yml up -d