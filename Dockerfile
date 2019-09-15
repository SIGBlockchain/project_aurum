FROM golang:alpine

ARG BRANCH=master

RUN rm -rf /var/cache/apk/* && \
    apk update && \
    apk add --no-cache git gcc musl-dev && \
    go get github.com/SIGBlockchain/project_aurum/... && \
    cd /$GOPATH/src/github.com/SIGBlockchain/project_aurum/cmd/ && \
    git checkout $BRANCH && \
    git pull && \
    apk del git gcc musl-dev && \
    go build -o main

EXPOSE 62000

WORKDIR /$GOPATH/src/github.com/SIGBlockchain/project_aurum/cmd/

# CMD [ "./main", "-d", "-g"] 
