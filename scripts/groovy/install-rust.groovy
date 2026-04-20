#!/usr/bin/env groovy

def call() {
    sh '''
        CARGO_HOME="/opt/rust"
        RUSTUP_HOME="/opt/rustup"
        export CARGO_HOME RUSTUP_HOME
        export PATH=$PATH:${CARGO_HOME}/bin

        if ! command -v cargo >/dev/null 2>&1; then
            curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs \
                | sudo env CARGO_HOME=$CARGO_HOME RUSTUP_HOME=$RUSTUP_HOME \
                  sh -s -- -y --default-toolchain stable --no-modify-path
            sudo ln -sf ${CARGO_HOME}/bin/cargo  /usr/local/bin/cargo
            sudo ln -sf ${CARGO_HOME}/bin/rustc  /usr/local/bin/rustc
            sudo ln -sf ${CARGO_HOME}/bin/rustup /usr/local/bin/rustup
        fi

        # Ensure stable toolchain is set as default
        sudo env CARGO_HOME=$CARGO_HOME RUSTUP_HOME=$RUSTUP_HOME \
            ${CARGO_HOME}/bin/rustup default stable

        sudo tee /etc/profile.d/rust.sh > /dev/null << EOF
export CARGO_HOME=${CARGO_HOME}
export RUSTUP_HOME=${RUSTUP_HOME}
export PATH=\\$PATH:\\$CARGO_HOME/bin
EOF
        sudo chmod 644 /etc/profile.d/rust.sh

        echo "CARGO_HOME=${CARGO_HOME}"
        RUSTUP_HOME=$RUSTUP_HOME ${CARGO_HOME}/bin/rustc --version
        RUSTUP_HOME=$RUSTUP_HOME ${CARGO_HOME}/bin/cargo --version
    '''
}

return this
