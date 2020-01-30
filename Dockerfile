FROM golang:latest

ARG PORT=26000

ENV ROOT=/go/src/github.com/SIGBlockchain/project_aurum

COPY . ${ROOT}

WORKDIR ${ROOT}/bin

EXPOSE ${PORT}

ENTRYPOINT ./main
