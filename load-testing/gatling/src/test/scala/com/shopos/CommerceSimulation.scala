package com.shopos

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import scala.concurrent.duration._

/**
 * Commerce simulation: full purchase journey.
 * login → browse → add to cart → checkout → payment
 */
class CommerceSimulation extends BaseSimulation {

  // ── Auth ──────────────────────────────────────────────────────────────────
  val loginScenario = scenario("Login")
    .exec(
      http("POST /auth/login")
        .post("/api/v1/auth/login")
        .body(StringBody("""{"email":"user1@test.shopos.local","password":"Password1!"}"""))
        .check(okStatus)
        .check(jsonPath("$.access_token").saveAs("authToken"))
    )
    .pause(300.milliseconds, 800.milliseconds)

  // ── Cart ──────────────────────────────────────────────────────────────────
  val createCartScenario = scenario("Create Cart")
    .exec(loginScenario)
    .exec(
      http("POST /cart")
        .post("/api/v1/cart")
        .header("Authorization", "Bearer #{authToken}")
        .body(StringBody("""{"userId":"load-test-user"}"""))
        .check(okOrCreated)
        .check(jsonPath("$.cartId").saveAs("cartId"))
    )
    .pause(200.milliseconds, 600.milliseconds)

  // ── Add to Cart ───────────────────────────────────────────────────────────
  val addToCartScenario = exec(
    feed(productIdFeeder),
    http("POST /cart/{id}/items")
      .post("/api/v1/cart/#{cartId}/items")
      .header("Authorization", "Bearer #{authToken}")
      .body(StringBody("""{"productId":"#{productId}","quantity":1}"""))
      .check(okOrCreated)
  ).pause(400.milliseconds, 1.second)

  // ── Checkout ──────────────────────────────────────────────────────────────
  val checkoutScenario = exec(
    http("POST /checkout")
      .post("/api/v1/checkout")
      .header("Authorization", "Bearer #{authToken}")
      .body(StringBody(
        """{
          |  "cartId":"#{cartId}",
          |  "shippingAddress":{"line1":"123 Test St","city":"New York","state":"NY","zip":"10001","country":"US"},
          |  "shippingMethod":"standard"
          |}""".stripMargin
      ))
      .check(okOrCreated)
      .check(jsonPath("$.orderId").saveAs("orderId"))
  ).pause(200.milliseconds, 500.milliseconds)

  // ── Payment ───────────────────────────────────────────────────────────────
  val paymentScenario = exec(
    http("POST /payments")
      .post("/api/v1/payments")
      .header("Authorization", "Bearer #{authToken}")
      .body(StringBody(
        """{
          |  "orderId":"#{orderId}",
          |  "paymentMethod":{"type":"card","number":"4111111111111111","expiry":"12/28","cvv":"123"},
          |  "amount":{"value":4999,"currency":"USD"}
          |}""".stripMargin
      ))
      .check(okOrCreated)
  )

  // ── Full journey ──────────────────────────────────────────────────────────
  val fullPurchaseJourney = scenario("Full Purchase Journey")
    .exec(loginScenario)
    .exec(
      http("GET /products")
        .get("/api/v1/products?page=1&size=20")
        .header("Authorization", "Bearer #{authToken}")
        .check(okStatus)
    )
    .pause(500.milliseconds, 2.seconds)
    .feed(productIdFeeder)
    .exec(
      http("GET /products/{id}")
        .get("/api/v1/products/#{productId}")
        .header("Authorization", "Bearer #{authToken}")
        .check(okStatus)
    )
    .pause(1.second, 3.seconds)
    .exec(
      http("POST /cart")
        .post("/api/v1/cart")
        .header("Authorization", "Bearer #{authToken}")
        .body(StringBody("""{"userId":"load-test-user"}"""))
        .check(okOrCreated)
        .check(jsonPath("$.cartId").saveAs("cartId"))
    )
    .exec(addToCartScenario)
    .exec(checkoutScenario)
    .exec(paymentScenario)
    .pause(1.second, 3.seconds)

  setUp(
    fullPurchaseJourney.inject(
      nothingFor(5.seconds),
      atOnceUsers(5),
      rampUsers(20) during (1.minute),
      constantUsersPerSec(5) during (5.minutes),
    )
  )
    .protocols(httpProtocol)
    .assertions(globalAssertions: _*)
    .assertions(
      details("POST /payments").responseTime.percentile(95).lt(5000),
      details("POST /checkout").responseTime.percentile(95).lt(3000),
    )
}
