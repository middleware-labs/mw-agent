#!/bin/bash
if [ -z "$1" ]
then
      echo "please pass target as 1st argument"
      exit
fi
if [ -z "$2" ]
then
      echo "please pass release tag as 2nd argument"
      exit
fi
docker build . --target $1 -t ghcr.io/middleware-labs/mw-agent:$2 -f $3

