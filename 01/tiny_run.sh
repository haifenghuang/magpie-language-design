#!/usr/bin/env bash
export GOPATH=$(pwd)

# for newer go version's build error: "package XXX is not in GOROOT"
export GO111MODULE=off

interpreter_name=magpie

tinygo build -o magpie.exe main.go
