FROM golang:alpine

RUN apk add --no-cache git gcc musl-dev \
    && go get github.com/SIGBlockchain/project_aurum/... \
    && apk del git gcc musl-dev


RUN cd /$GOPATH/src/github.com/SIGBlockchain/project_aurum/producer/src/main && go build -o main

EXPOSE 13131

WORKDIR /$GOPATH/src/github.com/SIGBlockchain/project_aurum/producer/src/main

CMD [ "./main" ] 
