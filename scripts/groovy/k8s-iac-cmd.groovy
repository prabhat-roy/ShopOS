#!/usr/bin/env groovy

def call() {
    // Returns the correct CLI command and maps TF_DIR to OpenTofu dir if needed
    def tool = env.IaC_TOOL ?: 'terraform'
    return tool  // 'terraform' or 'tofu'
}

return this
