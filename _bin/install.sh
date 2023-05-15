#!/bin/bash

BINARY="gomock"
USER_BIN=$HOME/bin
OS="$1"
ARCH="$2"

if [ -z "$OS" ]; then
    OS="darwin"
    echo "setting OS to ${OS}"
fi

if [ -z "$ARCH" ]; then
    ARCH="amd64"
    echo "setting ARCH to ${ARCH}"
fi

echo "installing ${BINARY} for ${OS} ${ARCH}"

link=$(curl -s "https://api.github.com/repos/smeshkov/gomock/releases/latest" | grep "browser_download_url.*${BINARY}_${OS}_${ARCH}" | cut -d : -f 2,3 | tr -d \" | tr -d ' ')
if [ -z "$link" ]; then
    echo "can't find ${BINARY} binary"
    exit 1
fi

echo "downloading ${BINARY} from $link"

curl -L -o ${BINARY} ${link}
chmod +x ${BINARY}

if [ ! -d "$USER_BIN" ]; then
  mkdir -p ${USER_BIN}
  echo "created $USER_BIN directory, don't forget to add it to PATH environment variable"
fi

echo "moving ${BINARY} to ${USER_BIN}/${BINARY}"

mv ${BINARY} ${USER_BIN}/${BINARY}

echo "installation is done."