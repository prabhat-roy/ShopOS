def call() {
    def language = env.LANGUAGE ?: 'go'
    def domain   = env.BUILD_DOMAIN

    sh 'mkdir -p reports/sca'

    // ── Trivy filesystem ──────────────────────────────────────────────────────
    sh """
        echo "=== SCA: Trivy filesystem ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            aquasec/trivy:latest fs \
            --format json \
            --output /src/reports/sca/trivy-fs.json \
            /src/src/${domain} || true
    """

    // ── Grype filesystem ──────────────────────────────────────────────────────
    sh """
        echo "=== SCA: Grype filesystem ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            anchore/grype:latest dir:/src/src/${domain} \
            --output json \
            > reports/sca/grype-fs.json 2>&1 || true
    """

    // ── Syft SBOM — SPDX + CycloneDX ─────────────────────────────────────────
    sh """
        echo "=== SCA: Syft SBOM (SPDX) ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            anchore/syft:latest dir:/src/src/${domain} \
            --output spdx-json \
            > reports/sca/sbom-spdx.json 2>&1 || true

        echo "=== SCA: Syft SBOM (CycloneDX) ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            anchore/syft:latest dir:/src/src/${domain} \
            --output cyclonedx-json \
            > reports/sca/sbom-cyclonedx.json 2>&1 || true
    """

    // ── OWASP Dependency-Check ────────────────────────────────────────────────
    sh """
        echo "=== SCA: OWASP Dependency-Check ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            -v \${WORKSPACE}/reports/sca:/report \
            owasp/dependency-check:latest \
            --project shopos-${domain} \
            --scan /src/src/${domain} \
            --format JSON \
            --out /report/dependency-check.json || true
    """

    // ── Snyk — open source dependencies ──────────────────────────────────────
    sh """
        echo "=== SCA: Snyk open source ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            -e SNYK_TOKEN=\${SNYK_TOKEN:-} \
            snyk/snyk:latest snyk test \
            /src/src/${domain} \
            --json > reports/sca/snyk-oss.json 2>&1 || true
    """

    // ── Docker Scout — supply chain (FS mode) ─────────────────────────────────
    sh """
        echo "=== SCA: Docker Scout (filesystem) ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            -v /var/run/docker.sock:/var/run/docker.sock \
            docker/scout-cli:latest cves \
            fs:///src/src/${domain} \
            --format sarif \
            --output /src/reports/sca/docker-scout-fs.json || true
    """

    // ── Vuls — OS + library vulnerability scan ────────────────────────────────
    sh """
        echo "=== SCA: Vuls ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            vuls/vuls:latest scan \
            -config /src/infra/vuls/config.toml \
            > reports/sca/vuls.log 2>&1 || true
    """

    // ── OpenSCAP — SCAP compliance scan ──────────────────────────────────────
    sh """
        echo "=== SCA: OpenSCAP ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            openscap/openscap:latest oscap \
            xccdf eval \
            --results /src/reports/sca/openscap-results.xml \
            --report /src/reports/sca/openscap-report.html \
            /usr/share/xml/scap/ssg/content/ssg-ubuntu2204-xccdf.xml \
            2>/dev/null || true
    """

    // ── Language-specific SCA ─────────────────────────────────────────────────
    if (language == 'go') {
        sh """
            echo "=== SCA: govulncheck ==="
            find \${WORKSPACE}/src/${domain} -name 'go.mod' -maxdepth 3 | while read gomod; do
                dir=\$(dirname "\$gomod")
                svc=\$(basename "\$dir")
                docker run --rm -v "\$dir":/app golang:1.23-alpine sh -c \
                    'go install golang.org/x/vuln/cmd/govulncheck@latest 2>/dev/null; \
                    cd /app && govulncheck -json ./... > /app/govulncheck.json 2>&1 || true'
                cp "\$dir/govulncheck.json" reports/sca/govulncheck-\$svc.json 2>/dev/null || true
            done
        """
    }

    if (language == 'nodejs') {
        sh """
            echo "=== SCA: npm audit ==="
            find \${WORKSPACE}/src/${domain} -name 'package.json' \
                -not -path '*/node_modules/*' -maxdepth 4 | while read pkg; do
                dir=\$(dirname "\$pkg")
                svc=\$(basename "\$dir")
                docker run --rm -v "\$dir":/app node:22-alpine sh -c \
                    'cd /app && npm install --prefer-offline -q 2>/dev/null; \
                    npm audit --json > /app/npm-audit.json 2>&1 || true'
                cp "\$dir/npm-audit.json" reports/sca/npm-audit-\$svc.json 2>/dev/null || true
            done
        """
    }

    if (language == 'python') {
        sh """
            echo "=== SCA: pip-audit ==="
            find \${WORKSPACE}/src/${domain} -name 'requirements.txt' -maxdepth 4 | while read req; do
                dir=\$(dirname "\$req")
                svc=\$(basename "\$dir")
                docker run --rm -v "\$dir":/app python:3.13-slim sh -c \
                    'pip install pip-audit -q 2>/dev/null && \
                    pip-audit -r /app/requirements.txt --format=json \
                    -o /app/pip-audit.json 2>/dev/null || true'
                cp "\$dir/pip-audit.json" reports/sca/pip-audit-\$svc.json 2>/dev/null || true
            done
        """
    }

    if (language == 'java' || language == 'kotlin') {
        sh """
            echo "=== SCA: Maven OWASP DC ==="
            find \${WORKSPACE}/src/${domain} -name 'pom.xml' -maxdepth 3 | while read pom; do
                dir=\$(dirname "\$pom")
                svc=\$(basename "\$dir")
                docker run --rm -v "\$dir":/app -v \${WORKSPACE}/reports/sca:/report \
                    maven:3.9-eclipse-temurin-21 sh -c \
                    'cd /app && mvn org.owasp:dependency-check-maven:check \
                    -Dformat=JSON -DoutputDirectory=/report \
                    --batch-mode -q 2>/dev/null || true' || true
            done
        """
    }

    if (language == 'rust') {
        sh """
            echo "=== SCA: cargo audit ==="
            find \${WORKSPACE}/src/${domain} -name 'Cargo.toml' -maxdepth 3 | while read cargo; do
                dir=\$(dirname "\$cargo")
                svc=\$(basename "\$dir")
                docker run --rm -v "\$dir":/app rust:1.81-slim sh -c \
                    'cargo install cargo-audit -q 2>/dev/null; \
                    cargo audit --json > /app/cargo-audit.json 2>&1 || true'
                cp "\$dir/cargo-audit.json" reports/sca/cargo-audit-\$svc.json 2>/dev/null || true
            done
        """
    }

    echo 'SCA & SBOM complete — reports/sca/'
}
return this
