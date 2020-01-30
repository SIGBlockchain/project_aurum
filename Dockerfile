FROM golang:latest

ARG PORT=26000

ENV ROOT=/go/src/github.com/SIGBlockchain/project_aurum

COPY . ${ROOT}

WORKDIR ${ROOT}/bin

EXPOSE ${PORT}

<<<<<<< HEAD
ENTRYPOINT ./main
=======
ENTRYPOINT [ "./main" ]
>>>>>>> 1597cdfe8aef71688c9609577636786845c92af1
