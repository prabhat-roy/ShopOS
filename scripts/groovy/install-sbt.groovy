#!/usr/bin/env groovy

def call() {
    sh '''
        if ! command -v sbt >/dev/null 2>&1; then
            echo "deb https://repo.scala-sbt.org/scalasbt/debian all main" \
                | sudo tee /etc/apt/sources.list.d/sbt.list > /dev/null
            curl -sL "https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x2EE0EA64E40A89B84B2DF73499E82A75642AC823" \
                | sudo apt-key add - 2>/dev/null
            sudo apt-get update -y
            sudo apt-get install -y sbt
        fi

        SBT_HOME=$(dirname $(dirname $(readlink -f $(which sbt))))

        sudo tee /etc/profile.d/sbt.sh > /dev/null << EOF
export SBT_HOME=${SBT_HOME}
export PATH=\\$PATH:\\$SBT_HOME/bin
EOF
        sudo chmod 644 /etc/profile.d/sbt.sh

        echo "SBT_HOME=${SBT_HOME}"
        sbt --version 2>/dev/null | head -1
    '''
}

return this
