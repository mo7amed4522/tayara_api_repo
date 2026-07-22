#!/usr/bin/env bash

opt=$1

if [ "$opt" == "" ];then

    if [ "$(uname)" == "Darwin" ];then
        echo "build use mac os" # Mac OS X Operating System
        opt='linux'

    elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ];then
        echo "build use linux os" # GNU/Linux Operating System
        opt='linux'

    elif [ "$(expr substr $(uname -s) 1 10)" == "MINGW64_NT" ];then
        echo "build use windows os" # Windows NT Operating System
        opt='windows'
    fi
fi

case "$opt" in
    'windows')
    target=server.exe
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.COMPILETIME=`date '+%Y-%m-%d_%H:%M:%S'` -X main.GITHASH=`git rev-parse HEAD`" -o $target main.go
        ;;
    'linux')
    target=server
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.COMPILETIME=`date '+%Y-%m-%d_%H:%M:%S'` -X main.GITHASH=`git rev-parse HEAD`" -o $target main.go
        ;;
        *)
        echo "Usage:./build.sh {windows|linux}"
        exit 1
        ;;
esac

exit 0