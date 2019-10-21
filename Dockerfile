FROM golang:alpine
ARG PORT
ARG INTERVAL=1h
ARG LOCALHOST=false
ARG GENESIS
ARG BRANCH

ENV PROJECT_ROOT=/go/src/github.com/SIGBlockchain/project_aurum

RUN rm -rf /var/cache/apk/* && \
    apk update && \
    apk add --no-cache git gcc musl-dev bash && \
    go get -v github.com/SIGBlockchain/project_aurum/... && \
    cd ${PROJECT_ROOT} && \
    git checkout ${BRANCH} && \
    git pull && \
    cd ${PROJECT_ROOT}/cmd && \
    go build -o main && \
    if [[ -n '${GENESIS}' ]]; then echo -e ${GENESIS} | \
    sed 's/ /\n/g' > ${PROJECT_ROOT}/cmd/genesis_hashes.txt; fi && \
    cd ${PROJECT_ROOT}/cmd/settings && \
    go run main.go -port ${PORT} -interval ${INTERVAL} -localhost=${LOCALHOST} && \    
    apk del git gcc musl-dev bash

EXPOSE ${PORT}

WORKDIR ${PROJECT_ROOT}/cmd

ENTRYPOINT [ "./main" ]