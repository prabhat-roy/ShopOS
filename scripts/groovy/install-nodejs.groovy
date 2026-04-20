#!/usr/bin/env groovy

def call() {
    sh '''
        if ! command -v node >/dev/null 2>&1; then
            curl -fsSL https://deb.nodesource.com/setup_22.x -o /tmp/nodesource-setup.sh
            sudo -E bash /tmp/nodesource-setup.sh
            rm -f /tmp/nodesource-setup.sh
            sudo apt-get install -y nodejs
        fi

        NODE_HOME=$(dirname $(dirname $(readlink -f $(which node))))

        sudo tee /etc/profile.d/nodejs.sh > /dev/null << EOF
export NODE_HOME=${NODE_HOME}
export PATH=\\$PATH:\\$NODE_HOME/bin
EOF
        sudo chmod 644 /etc/profile.d/nodejs.sh

        echo "NODE_HOME=${NODE_HOME}"
        node --version
        npm --version
    '''
}

return this
