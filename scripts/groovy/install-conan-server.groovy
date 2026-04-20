def call() {
    sh """
        helm upgrade --install conan-server registry/charts/conan-server \
            --namespace conan-server \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://conan-server-conan-server.conan-server.svc.cluster.local:9300'
    sh "sed -i '/^CONAN_SERVER_/d' infra.env || true"
    sh "sed -i '/^CONAN_SERVER_URL=/d' infra.env 2>/dev/null || true; echo 'CONAN_SERVER_URL=http://conan-server-conan-server.conan-server.svc.cluster.local:9300' >> infra.env" 
    sh "sed -i '/^CONAN_SERVER_USER=/d' infra.env 2>/dev/null || true; echo 'CONAN_SERVER_USER=demo' >> infra.env" 
    sh "sed -i '/^CONAN_SERVER_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'CONAN_SERVER_PASSWORD=demo' >> infra.env" 

    echo 'conan-server installed — ${url}'
}

return this
