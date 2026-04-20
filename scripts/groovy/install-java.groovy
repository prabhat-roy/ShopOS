#!/usr/bin/env groovy

def call() {
    sh '''
        if ! command -v java >/dev/null 2>&1; then
            sudo apt-get update -y
            sudo apt-get install -y fontconfig openjdk-21-jdk
        fi

        JAVA_HOME=$(dirname $(dirname $(readlink -f $(which java))))

        sudo tee /etc/profile.d/java.sh > /dev/null << EOF
export JAVA_HOME=${JAVA_HOME}
export PATH=\\$PATH:\\$JAVA_HOME/bin
EOF
        sudo chmod 644 /etc/profile.d/java.sh

        echo "JAVA_HOME=${JAVA_HOME}"
        java -version
    '''
}

return this
