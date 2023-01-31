#!/bin/sh -e

if ! command -v golangci-lint &> /dev/null
then
    echo "golangci-lint could not be found locally, visit https://golangci-lint.run/usage/install/"
    exit
fi

golangci-lint run