def call() {
    def language = env.LANGUAGE ?: 'go'
    def domain   = env.BUILD_DOMAIN

    env.SERVICES.split(',').each { svc ->
        svc = svc.trim()
        def svcPath = "src/${domain}/${svc}"
        echo "=== Compiling ${svc} (${language}) ==="

        switch (language) {

            case 'go':
                sh """
                    cd ${svcPath}
                    go mod download
                    go build ./...
                    go vet ./...
                    go test ./... -count=1 -timeout=60s 2>/dev/null || true
                """
                break

            case 'java':
                sh """
                    cd ${svcPath}
                    mvn clean package --batch-mode -q
                """
                break

            case 'kotlin':
                sh """
                    cd ${svcPath}
                    gradle build --no-daemon -q
                """
                break

            case 'python':
                sh """
                    cd ${svcPath}
                    pip install -r requirements.txt -q
                    python -m py_compile \$(find . -name '*.py' | head -20) || true
                """
                break

            case 'nodejs':
                sh """
                    cd ${svcPath}
                    npm ci --prefer-offline
                    if grep -q '"build"' package.json 2>/dev/null; then npm run build; fi
                    if grep -q '"test"' package.json 2>/dev/null; then npm test -- --passWithNoTests 2>/dev/null || true; fi
                """
                break

            case 'csharp':
                sh """
                    cd ${svcPath}
                    dotnet restore -q
                    dotnet build --configuration Release --no-restore -q
                    dotnet test --configuration Release --no-build --logger trx 2>/dev/null || true
                """
                break

            case 'rust':
                sh """
                    cd ${svcPath}
                    cargo build --release
                    cargo test 2>/dev/null || true
                """
                break

            case 'scala':
                sh """
                    cd ${svcPath}
                    sbt assembly -q
                """
                break

            default:
                echo "Unknown language '${language}' for ${svc} — skipping compile"
        }

        echo "${svc} compile done"
    }
}
return this
