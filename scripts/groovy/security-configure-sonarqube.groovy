def call() {
    sh """
        SONAR_URL=\$(grep '^SONARQUBE_URL=' infra.env | cut -d= -f2)
        echo "Waiting for SonarQube at \${SONAR_URL}..."
        until curl -sf "\${SONAR_URL}/api/system/status" | grep -q '"status":"UP"'; do sleep 10; done

        # Change default admin password
        curl -sf -u admin:admin -X POST "\${SONAR_URL}/api/users/change_password" \
            -d "login=admin&previousPassword=admin&password=admin123" || true

        # Create default project
        curl -sf -u admin:admin123 -X POST "\${SONAR_URL}/api/projects/create" \
            -d "name=ShopOS&project=shopos&visibility=private" || true

        # Generate user token
        TOKEN=\$(curl -sf -u admin:admin123 -X POST "\${SONAR_URL}/api/user_tokens/generate" \
            -d "name=jenkins-token&login=admin" | grep -o '"token":"[^"]*"' | cut -d: -f2 | tr -d '"')

        sed -i '/^SONARQUBE_TOKEN=/d' infra.env || true
        echo "SONARQUBE_TOKEN=\${TOKEN}" >> infra.env

        # Set quality gate — activate default gate
        curl -sf -u admin:admin123 -X POST "\${SONAR_URL}/api/qualitygates/set_as_default" \
            -d "name=Sonar+way" || true
    """
    echo 'sonarqube configured'
}
return this
