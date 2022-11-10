#!/usr/bin/env bash

OLD_PWD=$PWD
for path in `seq -w 1 53`
do
  cd $OLD_PWD/$path
  echo "clean dir $path ..."
  rm -rf magpie-*
  rm -f *.exe
done





