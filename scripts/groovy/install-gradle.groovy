#!/usr/bin/env groovy

def call() {
    sh '''
        GRADLE_VERSION="8.13"
        GRADLE_HOME="/opt/gradle/gradle-${GRADLE_VERSION}"

        if ! command -v gradle >/dev/null 2>&1; then
            curl -sL "https://services.gradle.org/distributions/gradle-${GRADLE_VERSION}-bin.zip" \
                -o /tmp/gradle.zip
            sudo unzip -oq /tmp/gradle.zip -d /opt/gradle
            sudo ln -sf ${GRADLE_HOME}/bin/gradle /usr/local/bin/gradle
            rm -f /tmp/gradle.zip
        fi

        sudo tee /etc/profile.d/gradle.sh > /dev/null << EOF
export GRADLE_HOME=${GRADLE_HOME}
export PATH=\\$PATH:\\$GRADLE_HOME/bin
EOF
        sudo chmod 644 /etc/profile.d/gradle.sh

        echo "GRADLE_HOME=${GRADLE_HOME}"
        gradle --version | grep Gradle
    '''
}

return this
