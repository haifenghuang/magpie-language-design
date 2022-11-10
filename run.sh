#!/usr/bin/env bash

OLD_PWD=$PWD
for path in `seq -w 1 53`
do
  cd $OLD_PWD/$path
  echo "In $path..."
  export GOPATH=$(pwd)
  # for newer go version's build error: "package XXX is not in GOROOT"
  export GO111MODULE=off

  interpreter_name=magpie

  platforms=("windows/amd64")
  for platform in "${platforms[@]}"
  do
      platform_split=(${platform//\// })
      GOOS=${platform_split[0]}
      GOARCH=${platform_split[1]}
      output_name=$interpreter_name'-'$GOOS'-'$GOARCH
      if [ $GOOS = "windows" ]; then
          output_name+='.exe'
      fi
  
      echo "    Building ${interpreter_name} ( $GOOS )...       ($output_name)"
      env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "-s -w" -o $output_name main.go
      if [ $? -ne 0 ]; then
          echo 'An error has occurred! Aborting the script execution...'
          exit 1
      fi
  done
done





