#!/usr/bin/env bash

Build(){
    echo -n "Building $1..."
    export GOOS=$2 GOARCH=$3 GO386=sse2
    go build -o "out/$1"
    echo Done!
}

# # OS X
# Build Tieba_Sign-Go.darwin darwin amd64

# # Windows
# Build Tieba_Sign-Go.x86.exe windows 386
# Build Tieba_Sign-Go.x64.exe windows amd64

# # Linux
Build Tieba_Sign-Go.linux.386.bin linux 386
# Build Tieba_Sign-Go.linux.amd64.bin linux amd64
# Build Tieba_Sign-Go.linux.arm.bin linux arm
# Build Tieba_Sign-Go.linux.arm64.bin linux arm64
