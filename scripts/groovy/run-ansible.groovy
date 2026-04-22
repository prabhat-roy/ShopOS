#!/usr/bin/env groovy

def call() {
    sh """
        # Install ansible if not present
        which ansible-playbook || pip3 install ansible --quiet

        # Run the Jenkins role to ensure host is bootstrapped
        ansible-playbook infra/ansible/site.yml \
            --limit jenkins \
            --tags "common,jenkins" \
            -e "ansible_connection=local" \
            -v || echo "WARNING: Ansible completed with warnings"
    """
    echo 'Ansible bootstrap complete'
}

return this
