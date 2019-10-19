# FROM golang:alpine
# ARG PORT=62000
# ARG WALLET_ADDRS

# ENV SRC_DIR=/go/src/github.com/SIGBlockchain/project_aurum

# RUN rm -rf /var/cache/apk/* && \
#     apk update && \
#     apk add --no-cache git gcc musl-dev bash && \
#     go get github.com/SIGBlockchain/project_aurum/... && \
#     cd $SRC_DIR/cmd/ && \
#     # git checkout $BRANCH && \
#     # git pull && \
#     go build -o main && \
#     bash -c "if [[ -n '${WALLET_ADDRS}' ]]; then echo -e ${WALLET_ADDRS} | sed 's/ /\n/g' > $SRC_DIR/cmd/genesis_hashes.txt; fi" && \
#     bash -c "cd $SRC_DIR/cmd/settings && go run main.go -port ${PORT}" && \
#     apk del git gcc musl-dev bash

# EXPOSE ${PORT}
# WORKDIR $SRC_DIR/cmd

# ENTRYPOINT ["./main"]

FROM golang:alpine
ARG PORT=62000
ARG INTERVAL=1h
ARG LOCALHOST=false
ARG GENESIS
ARG BRANCH=master

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