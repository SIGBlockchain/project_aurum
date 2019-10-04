#!/bin/bash
cd ../cmd/
go build -o main
expected=($(sha256sum main))
cd ../bin/
actual=($(sha256sum main))
echo "cmd/ binary sha256sum: $expected"
echo "bin/ binary sha256sum: $actual"
if [ "$expected" == "$actual" ]
then
    echo "Binaries match"
    exit 0
else
    echo "Binaries not updated"
    echo "Run rebuild script"
    exit 1
fi