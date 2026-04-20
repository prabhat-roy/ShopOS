#!/usr/bin/env bash
set -euo pipefail

# Install Buf CLI
BUF_VERSION=1.47.2
curl -sSL "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-Linux-x86_64" -o /usr/local/bin/buf
chmod +x /usr/local/bin/buf

# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Install k9s
K9S_VERSION=0.32.7
curl -sSL "https://github.com/derailed/k9s/releases/download/v${K9S_VERSION}/k9s_Linux_amd64.tar.gz" | tar -xz -C /usr/local/bin k9s

# Install Skaffold
curl -sSL https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 -o /usr/local/bin/skaffold
chmod +x /usr/local/bin/skaffold

# Install Tilt
curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash

# Python tooling
pip install --quiet pre-commit detect-secrets bandit

# Node global tools
npm install -g @bufbuild/protoc-gen-es @connectrpc/protoc-gen-connect-es

echo "ShopOS devcontainer ready."
