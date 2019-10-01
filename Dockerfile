FROM golang:alpine

ARG BRANCH=master
ARG WALLET_ADDR_1=23aafe84f813bd5093599691ea5731425effe1b8c3f7c1e3c049012558160b8c
ARG WALLET_ADDR_2=617af9da0ea3053e89ce43e1514640bcd41a3bd69810ca0f0d82e5cf8b4cadb8

RUN rm -rf /var/cache/apk/* && \
    apk update && \
    apk add --no-cache git gcc musl-dev && \
    go get github.com/SIGBlockchain/project_aurum/... && \
    cd /$GOPATH/src/github.com/SIGBlockchain/project_aurum/cmd/ && \
    git checkout $BRANCH && \
    git pull && \
    go build -o main && \
    apk del git gcc musl-dev && \
    echo $WALLET_ADDR_1 > genesis_hashes.txt && \
    echo $WALLET_ADDR_2 >> genesis_hashes.txt

EXPOSE 62000

WORKDIR /$GOPATH/src/github.com/SIGBlockchain/project_aurum/cmd/

# Change port in config.json && Run main
CMD cd settings/ && go run main.go -port 62000 && cd .. && ./main