FROM golang:alpine

ARG BRANCH=master

RUN rm -rf /var/cache/apk/* && \
    apk update && \
    apk add --no-cache git gcc musl-dev && \
    go get github.com/SIGBlockchain/project_aurum/... && \
    cd /$GOPATH/src/github.com/SIGBlockchain/project_aurum/cmd/ && \
    git checkout $BRANCH && \
    git pull && \
    go build -o main && \
    apk del git gcc musl-dev

EXPOSE 62000

WORKDIR /$GOPATH/src/github.com/SIGBlockchain/project_aurum/cmd/

# RUN rm -rf /var/cache/apk/* \
#     && rm -rf /tmp/* \
#     && apk update \
#     && apk add --no-cache git gcc musl-dev \
#     && go get github.com/SIGBlockchain/project_aurum/... \
#     && cd /$GOPATH/src/github.com/SIGBlockchain/project_aurum && git pull && git checkout Docker\
#     && apk del git gcc musl-dev


# RUN cd /$GOPATH/src/github.com/SIGBlockchain/project_aurum/internal/producer/main && go build -o main

# EXPOSE 13131

# WORKDIR /$GOPATH/src/github.com/SIGBlockchain/project_aurum/internal/producer/main

# CMD [ "./main", "-d", "-g"] 
