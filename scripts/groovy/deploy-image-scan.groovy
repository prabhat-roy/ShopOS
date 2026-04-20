def call() {
    sh 'mkdir -p reports/image-scan'

    // Login so scanners that pull from registry have credentials
    sh """
        echo "=== Registry login (image scan) ==="
        echo '${env.HARBOR_PASSWORD}' | \
            docker login ${env.HARBOR_URL} \
            -u ${env.HARBOR_USER} \
            --password-stdin || true
    """

    env.SERVICES.split(',').each { svc ->
        svc = svc.trim()
        def image = "${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${svc}:${env.IMAGE_TAG}"

        sh """
            echo "=== Image Scan: ${svc} ==="

            # Trivy — CVE + misconfig + secrets in image
            docker run --rm \
                -v /var/run/docker.sock:/var/run/docker.sock \
                aquasec/trivy:latest image \
                --format json \
                --exit-code 0 \
                --output /tmp/trivy-img-${svc}.json \
                ${image} || true
            cp /tmp/trivy-img-${svc}.json reports/image-scan/trivy-${svc}.json 2>/dev/null || true

            # Trivy — secrets scan on image layers
            docker run --rm \
                -v /var/run/docker.sock:/var/run/docker.sock \
                aquasec/trivy:latest image \
                --scanners secret \
                --format json \
                --output /tmp/trivy-secret-${svc}.json \
                ${image} || true
            cp /tmp/trivy-secret-${svc}.json reports/image-scan/trivy-secrets-${svc}.json 2>/dev/null || true

            # Trivy — CycloneDX SBOM from image
            docker run --rm \
                -v /var/run/docker.sock:/var/run/docker.sock \
                aquasec/trivy:latest image \
                --format cyclonedx \
                --output /tmp/sbom-img-${svc}.json \
                ${image} || true
            cp /tmp/sbom-img-${svc}.json reports/image-scan/sbom-image-${svc}.json 2>/dev/null || true

            # Grype — vulnerability scan
            docker run --rm \
                -v /var/run/docker.sock:/var/run/docker.sock \
                anchore/grype:latest ${image} \
                --output json \
                > reports/image-scan/grype-${svc}.json 2>&1 || true

            # Docker Scout — vulnerability and CVE scan
            docker run --rm \
                -v /var/run/docker.sock:/var/run/docker.sock \
                docker/scout-cli:latest cves ${image} \
                --format sarif \
                --output reports/image-scan/scout-${svc}.json || true

            # Syft — image SBOM (SPDX)
            docker run --rm \
                -v /var/run/docker.sock:/var/run/docker.sock \
                anchore/syft:latest ${image} \
                --output spdx-json \
                > reports/image-scan/syft-${svc}.json 2>&1 || true
        """

        // Anchore Engine — K8s server API
        if (env.ANCHORE_URL?.trim()) {
            sh """
                echo "=== Image Scan: Anchore Engine — ${svc} ==="
                # Submit image for analysis
                curl -sf -X POST "${env.ANCHORE_URL}/v1/images" \
                    -H "Content-Type: application/json" \
                    -u admin:\${ANCHORE_PASSWORD:-foobar} \
                    -d "{\"tag\":\"${image}\"}" || true
                # Wait and fetch vulns
                sleep 30
                curl -sf "${env.ANCHORE_URL}/v1/images/${image}/vuln/all" \
                    -u admin:\${ANCHORE_PASSWORD:-foobar} \
                    > reports/image-scan/anchore-${svc}.json 2>&1 || true
            """
        }

        // Clair — K8s server API
        if (env.CLAIR_URL?.trim()) {
            sh """
                echo "=== Image Scan: Clair — ${svc} ==="
                curl -sf -X POST "${env.CLAIR_URL}/indexer/api/v1/index_report" \
                    -H "Content-Type: application/json" \
                    -d "{\"hash\":\"${image}\"}" \
                    > reports/image-scan/clair-${svc}.json 2>&1 || true
            """
        }
    }

    echo 'Image scanning complete — reports/image-scan/'
}
return this
