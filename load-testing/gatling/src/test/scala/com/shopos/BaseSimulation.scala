package com.shopos

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import scala.concurrent.duration._

abstract class BaseSimulation extends Simulation {

  val baseUrl: String = sys.env.getOrElse("BASE_URL", "http://api-gateway.platform.svc.cluster.local:8080")

  val httpProtocol = http
    .baseUrl(baseUrl)
    .acceptHeader("application/json")
    .contentTypeHeader("application/json")
    .acceptEncodingHeader("gzip, deflate")
    .userAgentHeader("Gatling/ShopOS-LoadTest")
    .shareConnections

  // Common response checks
  val okStatus = status.is(200)
  val createdStatus = status.is(201)
  val okOrCreated = status.in(200, 201)

  // Feeder for product IDs
  val productIdFeeder = Iterator.continually(Map(
    "productId" -> Seq(
      "prod-001", "prod-002", "prod-003", "prod-004", "prod-005",
      "prod-010", "prod-020", "prod-030", "prod-040", "prod-050",
      "prod-100", "prod-200", "prod-300", "prod-400", "prod-500",
    )(scala.util.Random.nextInt(15))
  ))

  val categoryIdFeeder = Iterator.continually(Map(
    "categoryId" -> Seq(
      "cat-electronics", "cat-clothing", "cat-books",
      "cat-home", "cat-sports", "cat-beauty",
    )(scala.util.Random.nextInt(6))
  ))

  val searchQueryFeeder = Iterator.continually(Map(
    "searchQuery" -> Seq(
      "laptop", "smartphone", "headphones", "keyboard", "monitor",
      "shirt", "shoes", "jeans", "jacket", "dress",
      "python book", "coffee maker", "yoga mat", "running shoes",
    )(scala.util.Random.nextInt(14))
  ))

  val userFeeder = csv("users.csv").circular

  // SLO assertions applied to all simulations
  protected val globalAssertions = Seq(
    global.failedRequests.percent.lt(5.0),
    global.responseTime.percentile(95).lt(3000),
    global.responseTime.percentile(99).lt(5000),
  )
}
