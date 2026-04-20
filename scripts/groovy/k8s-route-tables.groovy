#!/usr/bin/env groovy

def call(String tfDir) {
    sh """
        cd ${tfDir}
        terraform apply \
            -target=aws_route_table.public \
            -target=aws_route_table.private \
            -target=aws_route_table_association.public \
            -target=aws_route_table_association.private \
            -auto-approve -input=false
    """
}

return this
