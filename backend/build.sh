#!/usr/bin/env bash
pushd sshbackend > /dev/null
echo "building sshbackend"
go build -ldflags="-s -w" -trimpath .
popd > /dev/null

pushd dummybackend > /dev/null
echo "building dummybackend"
go build -ldflags="-s -w" -trimpath .
popd > /dev/null

pushd externalbackendlauncher > /dev/null
echo "building externalbackendlauncher"
go build -ldflags="-s -w" -trimpath .
popd > /dev/null

if [ ! -d "sshappbackend/local-code/remote-bin" ]; then
  mkdir "sshappbackend/local-code/remote-bin"
fi

pushd sshappbackend/remote-code > /dev/null
echo "building sshappbackend/remote-code"
# Disable dynamic linking by disabling CGo.
# We need to make the remote code as generic as possible, so we do this
echo " - building for arm64"
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o ../local-code/remote-bin/rt-arm64 .
echo " - building for arm"
CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags="-s -w" -trimpath -o ../local-code/remote-bin/rt-arm .
echo " - building for amd64"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o ../local-code/remote-bin/rt-amd64 .
echo " - building for i386"
CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags="-s -w" -trimpath -o ../local-code/remote-bin/rt-386 .
popd > /dev/null

pushd sshappbackend/local-code > /dev/null
echo "building sshappbackend/local-code"
go build -ldflags="-s -w" -trimpath -o sshappbackend .
popd > /dev/null

pushd api > /dev/null
echo "building api"
go build -ldflags="-s -w" -trimpath .
popd > /dev/null
