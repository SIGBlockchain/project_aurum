FROM golang:alpine
ARG BRANCH=master
ARG PORT=62000
ARG WALLET_ADDRS
ARG INTERVAL="1h"
ARG LOCALHOST=false

ENV SRC_DIR=/go/src/github.com/SIGBlockchain/project_aurum

RUN rm -rf /var/cache/apk/* && \
    apk update && \
    apk add --no-cache git gcc musl-dev bash && \
    go get github.com/SIGBlockchain/project_aurum/... && \
    cd $SRC_DIR/cmd/ && \
    git checkout $BRANCH && \
    git pull && \
    go build -o main && \
    bash -c "if [[ -n '${WALLET_ADDRS}' ]]; then echo -e ${WALLET_ADDRS} | sed 's/ /\n/g' > $SRC_DIR/cmd/genesis_hashes.txt; fi" && \
    bash -c "cd $SRC_DIR/cmd/settings && go run main.go -port ${PORT} -interval ${INTERVAL} -localhost=${LOCALHOST}" && \
    apk del git gcc musl-dev bash

EXPOSE ${PORT}
WORKDIR $SRC_DIR/cmd

ENTRYPOINT ["./main"]