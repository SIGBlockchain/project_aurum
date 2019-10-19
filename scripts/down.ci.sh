#!/bin/bash
docker-compose -f ../docker-compose.ci.yml down
docker system prune -a -f