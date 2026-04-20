def call() {
    def language = env.LANGUAGE ?: 'go'
    def domain   = env.BUILD_DOMAIN

    sh 'mkdir -p reports/license'

    // ── FOSSA — license compliance for all languages ───────────────────────────
    sh """
        echo "=== License: FOSSA ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            -e FOSSA_API_KEY=\${FOSSA_API_KEY:-} \
            fossas/fossa-cli:latest analyze \
            --dir /src/src/${domain} \
            --output \
            > reports/license/fossa.json 2>&1 || true
    """

    // ── Tern — container image layer license analysis ─────────────────────────
    env.SERVICES.split(',').each { svc ->
        svc = svc.trim()
        def image = "${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${svc}:${env.IMAGE_TAG}"
        sh """
            echo "=== License: Tern — ${svc} ==="
            docker run --rm \
                -v /var/run/docker.sock:/var/run/docker.sock \
                philipssoftware/tern:latest report \
                -f json \
                -i ${image} \
                > reports/license/tern-${svc}.json 2>&1 || true
        """
    }

    // ── Node.js license-checker ───────────────────────────────────────────────
    if (language == 'nodejs') {
        sh """
            echo "=== License: license-checker (Node.js) ==="
            find \${WORKSPACE}/src/${domain} -name 'package.json' \
                -not -path '*/node_modules/*' -maxdepth 4 | while read pkg; do
                dir=\$(dirname "\$pkg")
                svc=\$(basename "\$dir")
                docker run --rm -v "\$dir":/app node:22-alpine sh -c \
                    'cd /app && npm install --prefer-offline -q 2>/dev/null; \
                    npx license-checker --json > /app/license-check.json 2>/dev/null || true'
                cp "\$dir/license-check.json" reports/license/npm-licenses-\$svc.json 2>/dev/null || true
            done
        """
    }

    // ── Python license scan ────────────────────────────────────────────────────
    if (language == 'python') {
        sh """
            echo "=== License: pip-licenses ==="
            find \${WORKSPACE}/src/${domain} -name 'requirements.txt' -maxdepth 4 | while read req; do
                dir=\$(dirname "\$req")
                svc=\$(basename "\$dir")
                docker run --rm -v "\$dir":/app python:3.13-slim sh -c \
                    'pip install pip-licenses -q 2>/dev/null && \
                    pip install -r /app/requirements.txt -q 2>/dev/null && \
                    pip-licenses --format=json > /app/pip-licenses.json 2>/dev/null || true'
                cp "\$dir/pip-licenses.json" reports/license/pip-licenses-\$svc.json 2>/dev/null || true
            done
        """
    }

    // ── Go license scan ────────────────────────────────────────────────────────
    if (language == 'go') {
        sh """
            echo "=== License: go-licenses ==="
            find \${WORKSPACE}/src/${domain} -name 'go.mod' -maxdepth 3 | while read gomod; do
                dir=\$(dirname "\$gomod")
                svc=\$(basename "\$dir")
                docker run --rm -v "\$dir":/app golang:1.23-alpine sh -c \
                    'go install github.com/google/go-licenses@latest 2>/dev/null; \
                    cd /app && go-licenses csv ./... > /app/go-licenses.csv 2>/dev/null || true'
                cp "\$dir/go-licenses.csv" reports/license/go-licenses-\$svc.csv 2>/dev/null || true
            done
        """
    }

    echo 'License compliance complete — reports/license/'
}
return this
