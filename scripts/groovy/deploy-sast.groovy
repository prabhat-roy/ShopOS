def call() {
    sh 'mkdir -p reports/sast'

    def language = env.LANGUAGE ?: 'go'
    def domain   = env.BUILD_DOMAIN
    def services = env.SERVICES.split(',')

    // ── Semgrep — multi-language, all code ───────────────────────────────────
    sh """
        echo "=== SAST: Semgrep ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            semgrep/semgrep:latest semgrep scan \
            --config=auto \
            --json \
            --output=/src/reports/sast/semgrep.json \
            /src/src/${domain} || true
    """

    // ── ShellCheck — all shell scripts in the repo ────────────────────────────
    sh """
        echo "=== SAST: ShellCheck ==="
        find \${WORKSPACE} -name '*.sh' -not -path '*/vendor/*' -not -path '*/.git/*' \
            > /tmp/sh-files.txt 2>/dev/null || true
        if [ -s /tmp/sh-files.txt ]; then
            docker run --rm \
                -v \${WORKSPACE}:/src \
                koalaman/shellcheck:stable \
                --format=json \$(cat /tmp/sh-files.txt | sed 's|'"${WORKSPACE}"'|/src|g') \
                > reports/sast/shellcheck.json 2>&1 || true
        fi
    """

    // ── Spectral — OpenAPI / AsyncAPI / JSON schema lint ─────────────────────
    sh """
        echo "=== SAST: Spectral (API spec linting) ==="
        find \${WORKSPACE}/proto -name '*.yaml' -o -name '*.yml' -o -name '*.json' \
            2>/dev/null | head -20 | while read spec; do
            docker run --rm \
                -v \${WORKSPACE}:/src \
                stoplight/spectral:latest lint "/src/\${spec#\${WORKSPACE}/}" \
                --format json 2>/dev/null \
                >> reports/sast/spectral.json || true
        done
    """

    // ── Go ────────────────────────────────────────────────────────────────────
    if (language == 'go') {
        sh """
            echo "=== SAST: GoSec ==="
            docker run --rm \
                -v \${WORKSPACE}:/src \
                securego/gosec:latest \
                -fmt=json -out=/src/reports/sast/gosec.json \
                /src/src/${domain}/... || true

            echo "=== SAST: GolangCI-Lint ==="
            docker run --rm \
                -v \${WORKSPACE}:/src \
                golangci/golangci-lint:latest \
                golangci-lint run /src/src/${domain}/... \
                --out-format json \
                > reports/sast/golangci.json 2>&1 || true
        """
    }

    // ── Java ──────────────────────────────────────────────────────────────────
    if (language == 'java') {
        sh """
            echo "=== SAST: SpotBugs ==="
            find \${WORKSPACE}/src/${domain} -name 'pom.xml' -maxdepth 3 | while read pom; do
                dir=\$(dirname "\$pom")
                svc=\$(basename "\$dir")
                docker run --rm -v "\$dir":/app -v \${WORKSPACE}/reports/sast:/report \
                    maven:3.9-eclipse-temurin-21 sh -c \
                    'cd /app && mvn com.github.spotbugs:spotbugs-maven-plugin:spotbugs \
                    -Dspotbugs.xmlOutput=true --batch-mode -q 2>/dev/null; \
                    cp target/spotbugsXml.xml /report/spotbugs-'\$svc'.xml 2>/dev/null || true' || true
            done

            echo "=== SAST: PMD ==="
            docker run --rm \
                -v \${WORKSPACE}:/src \
                pmdcode/pmd:latest pmd \
                -d /src/src/${domain} \
                -R rulesets/java/quickstart.xml \
                -f json -r /src/reports/sast/pmd.json || true
        """
    }

    // ── Kotlin ────────────────────────────────────────────────────────────────
    if (language == 'kotlin') {
        sh """
            echo "=== SAST: Detekt ==="
            find \${WORKSPACE}/src/${domain} -name 'build.gradle.kts' -maxdepth 3 | while read gradle; do
                dir=\$(dirname "\$gradle")
                svc=\$(basename "\$dir")
                docker run --rm -v "\$dir":/app \
                    gradle:8.10-jdk21 sh -c \
                    'cd /app && gradle detekt --no-daemon -q 2>/dev/null | tee /app/detekt-out.txt || true'
                cp "\$dir/detekt-out.txt" reports/sast/detekt-\$svc.txt 2>/dev/null || true
            done
        """
    }

    // ── Python ────────────────────────────────────────────────────────────────
    if (language == 'python') {
        sh """
            echo "=== SAST: Bandit ==="
            docker run --rm \
                -v \${WORKSPACE}:/src \
                pycqa/bandit:latest bandit \
                -r /src/src/${domain} \
                -f json -o /src/reports/sast/bandit.json || true

            echo "=== SAST: Pylint ==="
            docker run --rm \
                -v \${WORKSPACE}:/src \
                python:3.13-slim sh -c \
                'pip install pylint -q 2>/dev/null && pylint /src/src/${domain} \
                --output-format=json 2>/dev/null > /src/reports/sast/pylint.json || true'

            echo "=== SAST: Flake8 ==="
            docker run --rm \
                -v \${WORKSPACE}:/src \
                python:3.13-slim sh -c \
                'pip install flake8 -q 2>/dev/null && \
                flake8 /src/src/${domain} --format=default \
                > /src/reports/sast/flake8.txt 2>&1 || true'

            echo "=== SAST: Pyflakes ==="
            docker run --rm \
                -v \${WORKSPACE}:/src \
                python:3.13-slim sh -c \
                'pip install pyflakes -q 2>/dev/null && \
                python -m pyflakes /src/src/${domain} \
                > /src/reports/sast/pyflakes.txt 2>&1 || true'
        """
    }

    // ── Node.js ───────────────────────────────────────────────────────────────
    if (language == 'nodejs') {
        sh """
            echo "=== SAST: ESLint ==="
            docker run --rm \
                -v \${WORKSPACE}:/src \
                node:22-alpine sh -c \
                'cd /src && npm install --save-dev eslint @eslint/js -q 2>/dev/null; \
                npx eslint src/${domain} \
                --format=json --output-file=/src/reports/sast/eslint.json || true'

            echo "=== SAST: retire.js ==="
            docker run --rm \
                -v \${WORKSPACE}:/src \
                node:22-alpine sh -c \
                'npm install -g retire -q 2>/dev/null; \
                retire --path /src/src/${domain} \
                --outputformat json \
                --outputpath /src/reports/sast/retire.json || true'
        """
    }

    // ── C# ────────────────────────────────────────────────────────────────────
    if (language == 'csharp') {
        sh """
            echo "=== SAST: Roslyn Analyzers ==="
            find \${WORKSPACE}/src/${domain} -name '*.csproj' -maxdepth 3 | while read proj; do
                dir=\$(dirname "\$proj")
                svc=\$(basename "\$dir")
                docker run --rm -v "\$dir":/app \
                    mcr.microsoft.com/dotnet/sdk:8.0 sh -c \
                    'cd /app && dotnet build 2>&1 | tee /app/roslyn.txt || true'
                cp "\$dir/roslyn.txt" reports/sast/roslyn-\$svc.txt 2>/dev/null || true
            done
        """
    }

    // ── Rust ──────────────────────────────────────────────────────────────────
    if (language == 'rust') {
        sh """
            echo "=== SAST: cargo clippy ==="
            find \${WORKSPACE}/src/${domain} -name 'Cargo.toml' -maxdepth 3 | while read cargo; do
                dir=\$(dirname "\$cargo")
                svc=\$(basename "\$dir")
                docker run --rm -v "\$dir":/app \
                    rust:1.81-slim sh -c \
                    'cd /app && cargo clippy --message-format=json 2>/dev/null > /app/clippy.json || true'
                cp "\$dir/clippy.json" reports/sast/clippy-\$svc.json 2>/dev/null || true
            done
        """
    }

    // ── Scala ─────────────────────────────────────────────────────────────────
    if (language == 'scala') {
        sh """
            echo "=== SAST: Scalastyle ==="
            find \${WORKSPACE}/src/${domain} -name 'build.sbt' -maxdepth 3 | while read sbt; do
                dir=\$(dirname "\$sbt")
                svc=\$(basename "\$dir")
                docker run --rm -v "\$dir":/app \
                    sbtscala/scala-sbt:eclipse-temurin-17.0.8_1.9.6_3.3.1 sh -c \
                    'cd /app && sbt scalastyle 2>&1 | tee /app/scalastyle.txt || true'
                cp "\$dir/scalastyle.txt" reports/sast/scalastyle-\$svc.txt 2>/dev/null || true
            done
        """
    }

    // ── Snyk — code analysis (all languages) ──────────────────────────────────
    sh """
        echo "=== SAST: Snyk Code ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            -e SNYK_TOKEN=\${SNYK_TOKEN:-} \
            snyk/snyk:latest snyk code test \
            /src/src/${domain} \
            --json > reports/sast/snyk-code.json 2>&1 || true
    """

    // ── Brakeman — Ruby (runs even if not Ruby; skips gracefully) ─────────────
    sh """
        echo "=== SAST: Brakeman (Ruby/Rails) ==="
        find \${WORKSPACE}/src/${domain} -name 'Gemfile' -maxdepth 3 | while read gem; do
            dir=\$(dirname "\$gem")
            svc=\$(basename "\$dir")
            docker run --rm -v "\$dir":/app \
                presidentbeef/brakeman:latest \
                --output /app/brakeman.json --quiet || true
            cp "\$dir/brakeman.json" reports/sast/brakeman-\$svc.json 2>/dev/null || true
        done
    """

    // ── SonarQube — server-side analysis (K8s tool) ────────────────────────────
    if (env.SONAR_TOKEN?.trim()) {
        services.each { svc ->
            svc = svc.trim()
            sh """
                echo "=== SAST: SonarQube — ${svc} ==="
                docker run --rm \
                    -v \${WORKSPACE}:/src \
                    sonarsource/sonar-scanner-cli:latest \
                    sonar-scanner \
                    -Dsonar.projectKey=${svc} \
                    -Dsonar.projectName="${svc}" \
                    -Dsonar.sources=/src/src/${domain}/${svc} \
                    -Dsonar.host.url=${env.SONAR_URL} \
                    -Dsonar.token=${env.SONAR_TOKEN} || true
            """
        }
    } else {
        echo 'SONAR_TOKEN not set — skipping SonarQube'
    }

    echo 'SAST complete — reports/sast/'
}
return this
