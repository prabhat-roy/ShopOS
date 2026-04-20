#!/usr/bin/env groovy

def call() {
    sh '''
        if ! command -v dotnet >/dev/null 2>&1; then
            sudo apt-get update -y
            sudo apt-get install -y dotnet-sdk-8.0
        fi

        DOTNET_ROOT=$(dirname $(readlink -f $(which dotnet)))

        sudo tee /etc/profile.d/dotnet.sh > /dev/null << EOF
export DOTNET_ROOT=${DOTNET_ROOT}
export PATH=\\$PATH:\\$DOTNET_ROOT:\\$HOME/.dotnet/tools
EOF
        sudo chmod 644 /etc/profile.d/dotnet.sh

        echo "DOTNET_ROOT=${DOTNET_ROOT}"
        dotnet --version
    '''
}

return this
