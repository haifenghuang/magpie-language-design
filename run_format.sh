#!/usr/bin/env bash

OLD_PWD=$PWD

for path in {1..53}
do
  cd $OLD_PWD/$path
  echo "In $path..."
  export GOPATH=$(pwd)
  # for newer go version's build error: "package XXX is not in GOROOT"
  export GO111MODULE=off

  # format each go file
  echo "Formatting go file..."
  go fmt ./main.go > /dev/null
  echo "    main.go"
  
  for file in `find ./src/magpie -name "*.go"`; do
  	echo "    `basename $file`"
  	go fmt $file > /dev/null
  done
done
