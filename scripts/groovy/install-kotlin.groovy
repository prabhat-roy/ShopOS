#!/usr/bin/env groovy

def call() {
    sh '''
        KOTLIN_VERSION="2.1.20"
        KOTLIN_HOME="/opt/kotlin/kotlinc"

        if ! command -v kotlinc >/dev/null 2>&1; then
            curl -sLO "https://github.com/JetBrains/kotlin/releases/download/v${KOTLIN_VERSION}/kotlin-compiler-${KOTLIN_VERSION}.zip"
            sudo unzip -q -d /opt/kotlin "kotlin-compiler-${KOTLIN_VERSION}.zip"
            sudo ln -sf ${KOTLIN_HOME}/bin/kotlinc /usr/local/bin/kotlinc
            sudo ln -sf ${KOTLIN_HOME}/bin/kotlin   /usr/local/bin/kotlin
            rm -f "kotlin-compiler-${KOTLIN_VERSION}.zip"
        fi

        sudo tee /etc/profile.d/kotlin.sh > /dev/null << EOF
export KOTLIN_HOME=${KOTLIN_HOME}
export PATH=\\$PATH:\\$KOTLIN_HOME/bin
EOF
        sudo chmod 644 /etc/profile.d/kotlin.sh

        echo "KOTLIN_HOME=${KOTLIN_HOME}"
        kotlinc -version 2>&1
    '''
}

return this
