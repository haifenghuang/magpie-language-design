#!/usr/bin/env bash
export GOPATH=$(pwd)

export GO111MODULE=off

olddir=$(pwd)
# build-wasm
cd wasm/cmd/wasm
echo "Building wasm..."
tinygo build -o ../../assets/magpie.wasm -target wasm ./main.go

# run server
cd ../server
echo "Running server..."
echo "    Now open the browser, and type 'http://localhost:9090'"
tinygo run main.go


