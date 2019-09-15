FROM golang:alpine

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
