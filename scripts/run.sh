#!/bin/bash
docker run --rm -it --name producer -v ${PWD}/../data/:/go/src/github.com/SIGBlockchain/project_aurum/data/ test