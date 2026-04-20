#!/usr/bin/env groovy

def call(String tfDir) {
    sh """
        cd ${tfDir}
        terraform apply \
            -target=aws_iam_role.cluster \
            -target=aws_iam_role_policy_attachment.cluster \
            -target=aws_iam_role.node \
            -target=aws_iam_role_policy_attachment.node_minimal \
            -target=aws_iam_role_policy_attachment.node_ecr \
            -auto-approve -input=false
    """
}

return this
