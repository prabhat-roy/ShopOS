Feature: Health checks for ShopOS services

  Background:
    * def apiGatewayUrl = karate.properties['apiGatewayUrl'] || 'http://localhost:8080'
    * def webBffUrl = karate.properties['webBffUrl'] || 'http://localhost:8081'
    * def mobileBffUrl = karate.properties['mobileBffUrl'] || 'http://localhost:8082'

  Scenario: API Gateway health check
    Given url apiGatewayUrl
    When method get /healthz
    Then status 200
    And match response == { status: 'ok' }

  Scenario: Web BFF health check
    Given url webBffUrl
    When method get /healthz
    Then status 200
    And match response == { status: 'ok' }

  Scenario: Mobile BFF health check
    Given url mobileBffUrl
    When method get /healthz
    Then status 200
    And match response == { status: 'ok' }

  Scenario: API Gateway returns JSON on invalid route
    Given url apiGatewayUrl
    And path '/nonexistent-route-xyz'
    When method get
    Then status 404
    And match response.error != null

  Scenario: Product catalog service health
    * def catalogUrl = karate.properties['catalogUrl'] || 'http://localhost:50070'
    Given url catalogUrl
    When method get /healthz
    Then status 200

  Scenario: API Gateway metrics endpoint
    Given url apiGatewayUrl
    And path '/metrics'
    When method get
    Then status 200
    And match responseHeaders['content-type'][0] contains 'text/plain'
