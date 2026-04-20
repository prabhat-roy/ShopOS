def call() {
    sh """
        echo "Updating Nuclei templates..."
        kubectl exec -n nuclei deploy/nuclei-nuclei -- nuclei -update-templates || true

        # List available template counts
        TMPL_COUNT=\$(kubectl exec -n nuclei deploy/nuclei-nuclei -- \
            nuclei -list-templates 2>/dev/null | wc -l || echo 0)
        echo "Nuclei templates ready: \${TMPL_COUNT} templates loaded"
    """
    echo 'nuclei configured — templates updated'
}
return this
