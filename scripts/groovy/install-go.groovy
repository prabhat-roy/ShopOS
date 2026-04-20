#!/usr/bin/env groovy

def call() {
    sh '''
        GO_VERSION="1.24.2"
        GOROOT="/usr/local/go"

        if ! /usr/local/go/bin/go version >/dev/null 2>&1 && ! command -v go >/dev/null 2>&1; then
            curl -sL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz
            sudo rm -rf /usr/local/go
            sudo tar -C /usr/local -xzf /tmp/go.tar.gz
            rm -f /tmp/go.tar.gz
        fi

        sudo tee /etc/profile.d/go.sh > /dev/null << EOF
export GOROOT=${GOROOT}
export GOPATH=\\$HOME/go
export PATH=\\$PATH:\\$GOROOT/bin:\\$GOPATH/bin
EOF
        sudo chmod 644 /etc/profile.d/go.sh

        export PATH=$PATH:${GOROOT}/bin
        echo "GOROOT=${GOROOT}"
        go version
    '''
}

return this
