def call() {
    def svc    = env.TEST_SERVICE
    def url    = env.SERVICE_URL
    def domain = env.TEST_DOMAIN

    sh 'mkdir -p reports/load/gatling'

    def simulation = (domain == 'catalog') ? 'SearchSimulation' : 'CommerceSimulation'

    sh """
        mkdir -p /tmp/gatling/simulations/shopos

        cat > /tmp/gatling/simulations/shopos/CommerceSimulation.scala << 'SCEOF'
package shopos

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import scala.concurrent.duration._

class CommerceSimulation extends Simulation {

  val baseUrl   = System.getProperty("baseUrl", "http://localhost:8080")
  val userCount = System.getProperty("users", "100").toInt
  val duration  = System.getProperty("duration", "300").toInt.seconds

  val httpProtocol = http
    .baseUrl(baseUrl)
    .acceptHeader("application/json")
    .contentTypeHeader("application/json")

  val productIds = (1 to 100).map(i => f"prod-\$i%04d").toVector
  val feeder     = Iterator.continually(Map("productId" -> productIds(scala.util.Random.nextInt(productIds.size))))

  val browseCatalog = scenario("Browse Catalog")
    .feed(feeder)
    .exec(
      http("List Products")
        .get("/api/v1/products")
        .queryParam("page", "1")
        .queryParam("limit", "20")
        .check(status.is(200))
    )
    .pause(1, 3)
    .exec(
      http("Get Product")
        .get("/api/v1/products/#{productId}")
        .check(status.is(200))
    )

  val checkout = scenario("Checkout Flow")
    .exec(
      http("Login")
        .post("/api/v1/auth/login")
        .body(StringBody("""{"email":"load-user@shopos.io","password":"LoadTest123!"}"""))
        .check(status.is(200))
        .check(jsonPath("\$.access_token").saveAs("token"))
    )
    .pause(1)
    .exec(
      http("Add to Cart")
        .post("/api/v1/cart/items")
        .header("Authorization", "Bearer #{token}")
        .body(StringBody("""{"product_id":"prod-0001","quantity":1}"""))
        .check(status.in(200, 201))
    )
    .pause(1)
    .exec(
      http("Checkout")
        .post("/api/v1/checkout")
        .header("Authorization", "Bearer #{token}")
        .body(StringBody("""{"cart_id":"cart-001","shipping_address_id":"addr-001","payment_method_id":"pm-test"}"""))
        .check(status.in(200, 201))
    )

  setUp(
    browseCatalog.inject(rampUsers(userCount) during duration),
    checkout.inject(rampUsers(userCount / 10) during duration)
  ).protocols(httpProtocol)
}
SCEOF

        cat > /tmp/gatling/simulations/shopos/SearchSimulation.scala << 'SCEOF'
package shopos

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import scala.concurrent.duration._

class SearchSimulation extends Simulation {

  val baseUrl   = System.getProperty("baseUrl", "http://localhost:8080")
  val userCount = System.getProperty("users", "200").toInt
  val duration  = System.getProperty("duration", "300").toInt.seconds

  val httpProtocol = http
    .baseUrl(baseUrl)
    .acceptHeader("application/json")

  val queries = Array(
    "laptop", "gaming keyboard", "wireless headphones",
    "4k monitor", "mechanical keyboard", "smartphone",
    "tablet", "webcam", "microphone", "usb hub"
  )

  val searchFeeder = Iterator.continually {
    Map("query" -> queries(scala.util.Random.nextInt(queries.length)))
  }

  val searchScenario = scenario("Product Search")
    .feed(searchFeeder)
    .exec(
      http("Search Products")
        .get("/api/v1/products/search")
        .queryParam("q", "#{query}")
        .queryParam("page", "1")
        .queryParam("limit", "20")
        .check(status.is(200))
        .check(jsonPath("\$.total").exists)
    )
    .pause(500.milliseconds, 2.seconds)
    .exec(
      http("Get Suggestions")
        .get("/api/v1/products/suggest")
        .queryParam("prefix", "#{query}")
        .check(status.is(200))
    )

  setUp(
    searchScenario.inject(
      constantUsersPerSec(userCount / duration.toSeconds) during duration
    )
  ).protocols(httpProtocol)
    .assertions(
      global.responseTime.percentile3.lt(500),
      global.successfulRequests.percent.gte(99)
    )
}
SCEOF

        echo "=== Gatling: ${simulation} → ${url} ==="
        docker run --rm \
            --network host \
            -v /tmp/gatling/simulations:/opt/gatling/user-files/simulations \
            -v \${WORKSPACE}/reports/load/gatling:/opt/gatling/results \
            -e JAVA_OPTS="-DbaseUrl=${url}" \
            denvazh/gatling:latest \
            -s shopos.${simulation} \
            -rd "ShopOS ${svc} load test - build ${env.BUILD_NUMBER}" || true

        echo "Gatling complete — reports/load/gatling/"
        rm -rf /tmp/gatling
    """
}
return this
