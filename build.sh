#!/usr/bin/bash

set -e

mkdir -p "bin/linux_x86" "bin/linux_amd64"
GOOS=linux GOARCH=386 go build -o bin/linux_x86/go-rename
GOOS=linux GOARCH=amd64 go build -o bin/linux_amd64/go-rename

mkdir -p "bin/osx_x86" "bin/osx_amd64"
GOOS=darwin GOARCH=386 go build -o bin/osx_x86/go-rename
GOOS=darwin GOARCH=amd64 go build -o bin/osx_amd64/go-rename

mkdir -p "bin/windows_x86" "bin/windows_amd64"
GOOS=windows GOARCH=386 go build -o bin/windows_x86/go-rename.exe
GOOS=windows GOARCH=amd64 go build -o bin/windows_amd64/go-rename.exe

go install
