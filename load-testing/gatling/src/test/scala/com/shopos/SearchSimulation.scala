package com.shopos

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import scala.concurrent.duration._

/**
 * Search simulation — exercises the search-service under high concurrency.
 * keyword search → filtered → autocomplete → facets
 */
class SearchSimulation extends BaseSimulation {

  val keywordSearch = scenario("Keyword Search")
    .feed(searchQueryFeeder)
    .exec(
      http("GET /search")
        .get("/api/v1/search?q=#{searchQuery}&page=1&size=20")
        .check(okStatus)
        .check(responseTimeInMillis.lte(2000))
    )
    .pause(800.milliseconds, 2.seconds)

  val filteredSearch = scenario("Filtered Search")
    .feed(searchQueryFeeder)
    .feed(categoryIdFeeder)
    .exec(
      http("GET /search?filtered")
        .get("/api/v1/search?q=#{searchQuery}&category=#{categoryId}&price_max=500&sort=price_asc&page=1&size=20")
        .check(okStatus)
    )
    .pause(1.second, 2.5.seconds)

  val autocomplete = scenario("Autocomplete")
    .feed(searchQueryFeeder)
    .exec(session => session.set("partial", session("searchQuery").as[String].take(3)))
    .exec(
      http("GET /search/suggest")
        .get("/api/v1/search/suggest?q=#{partial}&limit=5")
        .check(okStatus)
        .check(responseTimeInMillis.lte(300))
    )
    .pause(80.milliseconds, 150.milliseconds)

  val facets = scenario("Search Facets")
    .feed(searchQueryFeeder)
    .exec(
      http("GET /search/facets")
        .get("/api/v1/search/facets?q=#{searchQuery}")
        .check(okStatus)
    )
    .pause(600.milliseconds, 1.5.seconds)

  setUp(
    keywordSearch.inject(
      rampUsers(40) during (1.minute),
      constantUsersPerSec(8) during (5.minutes),
    ),
    filteredSearch.inject(
      nothingFor(30.seconds),
      rampUsers(20) during (1.minute),
      constantUsersPerSec(4) during (5.minutes),
    ),
    autocomplete.inject(
      nothingFor(15.seconds),
      rampUsers(30) during (1.minute),
      constantUsersPerSec(10) during (5.minutes),
    ),
    facets.inject(
      nothingFor(1.minute),
      rampUsers(10) during (1.minute),
      constantUsersPerSec(2) during (5.minutes),
    ),
  )
    .protocols(httpProtocol)
    .assertions(
      global.failedRequests.percent.lt(1.0),
      global.responseTime.percentile(95).lt(2500),
      details("GET /search/suggest").responseTime.percentile(95).lt(300),
      details("GET /search").responseTime.percentile(95).lt(2000),
    )
}
