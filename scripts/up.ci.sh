#!/bin/bash
export TRAVISBRANCH="$1"
docker-compose -f ../docker-compose.ci.yml up -d