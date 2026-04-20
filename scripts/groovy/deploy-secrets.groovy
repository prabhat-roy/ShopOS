def call() {
    sh 'mkdir -p reports/secrets'

    // ── Gitleaks ──────────────────────────────────────────────────────────────
    sh """
        echo "=== Secrets: Gitleaks ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            zricethezav/gitleaks:latest detect \
            --source=/src \
            --report-path=/src/reports/secrets/gitleaks.json \
            --report-format=json \
            --exit-code=0 || true
    """

    // ── TruffleHog ────────────────────────────────────────────────────────────
    sh """
        echo "=== Secrets: TruffleHog ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            trufflesecurity/trufflehog:latest filesystem /src \
            --json 2>/dev/null > reports/secrets/trufflehog.json || true
    """

    // ── GitGuardian ───────────────────────────────────────────────────────────
    sh """
        echo "=== Secrets: GitGuardian ggshield ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            -e GITGUARDIAN_API_KEY=\${GITGUARDIAN_API_KEY:-dummy} \
            gitguardian/ggshield:latest secret scan path /src \
            --json --exit-zero \
            > reports/secrets/gitguardian.json 2>&1 || true
    """

    // ── detect-secrets (Yelp/IBM) ─────────────────────────────────────────────
    sh """
        echo "=== Secrets: detect-secrets ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            python:3.13-slim sh -c \
            'pip install detect-secrets -q 2>/dev/null && \
            detect-secrets scan /src/src \
            --all-files \
            --json > /src/reports/secrets/detect-secrets.json 2>/dev/null || true'
    """

    echo 'Secret scanning complete — reports/secrets/'
}
return this
