#!/bin/bash

cd "$(dirname "$0")"

case "$(uname -s)" in
Darwin)
  echo 'build on Mac OS X'
  ;;
Linux)
  echo 'build on Linux'
  ;;
CYGWIN* | MINGW* | MSYS*)
  echo 'build on MS Windows'
  exe_ext='.exe'
  ;;

*)
  echo 'Other OS'
  ;;
esac
rm -f ~/bin/protoc-gen-ygrpc-protodb${exe_ext}
go build -o ~/bin/protoc-gen-ygrpc-protodb${exe_ext} protoc-gen-ygrpc-protodb.go
