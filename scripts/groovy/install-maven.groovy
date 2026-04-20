#!/usr/bin/env groovy

def call() {
    sh '''
        if ! command -v mvn >/dev/null 2>&1; then
            sudo apt-get update -y
            sudo apt-get install -y maven
        fi

        MAVEN_HOME=$(mvn --version 2>/dev/null | grep "Maven home" | awk '{print $NF}')
        MAVEN_HOME=${MAVEN_HOME:-/usr/share/maven}

        sudo tee /etc/profile.d/maven.sh > /dev/null << EOF
export MAVEN_HOME=${MAVEN_HOME}
export M2_HOME=${MAVEN_HOME}
export PATH=\\$PATH:\\$MAVEN_HOME/bin
EOF
        sudo chmod 644 /etc/profile.d/maven.sh

        echo "MAVEN_HOME=${MAVEN_HOME}"
        mvn --version | head -1
    '''
}

return this
