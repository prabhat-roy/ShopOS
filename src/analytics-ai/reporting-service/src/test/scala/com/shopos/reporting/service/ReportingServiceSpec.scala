package com.shopos.reporting.service

import cats.effect.IO
import cats.effect.unsafe.implicits.global
import com.shopos.reporting.store.InMemoryReportStore
import org.scalatest.funsuite.AnyFunSuite
import org.scalatest.matchers.should.Matchers

class ReportingServiceSpec extends AnyFunSuite with Matchers {

  private def makeService(): (ReportingService, InMemoryReportStore) = {
    val store = InMemoryReportStore().unsafeRunSync()
    val svc   = ReportingService(store)
    (svc, store)
  }

  test("generateDailySales returns empty report when no data exists") {
    val (svc, _) = makeService()
    val report   = svc.generateDailySales("2024-01-01").unsafeRunSync()
    report.date           shouldBe "2024-01-01"
    report.totalRevenue   shouldBe 0.0
    report.orderCount     shouldBe 0
    report.avgOrderValue  shouldBe 0.0
    report.topProducts    shouldBe empty
  }

  test("generateUserMetrics returns empty metrics when no data exists") {
    val (svc, _) = makeService()
    val metrics  = svc.generateUserMetrics("2024-01-01").unsafeRunSync()
    metrics.date          shouldBe "2024-01-01"
    metrics.pageViews     shouldBe 0
    metrics.uniqueSessions shouldBe 0
    metrics.bounceRate    shouldBe 0.0
  }

  test("processEvent page.viewed increments page views") {
    val (svc, _) = makeService()
    val event    = """{"sessionId":"sess-1","pageUrl":"/home","timestamp":"2024-03-10T12:00:00Z"}"""
    svc.processEvent("analytics.page.viewed", event).unsafeRunSync()
    val metrics = svc.generateUserMetrics("2024-03-10").unsafeRunSync()
    metrics.pageViews      shouldBe 1
    metrics.uniqueSessions shouldBe 1
  }

  test("processEvent page.viewed counts duplicate sessions correctly") {
    val (svc, _) = makeService()
    val event1 = """{"sessionId":"sess-A","pageUrl":"/home","timestamp":"2024-03-10T12:00:00Z"}"""
    val event2 = """{"sessionId":"sess-A","pageUrl":"/products","timestamp":"2024-03-10T12:01:00Z"}"""
    val event3 = """{"sessionId":"sess-B","pageUrl":"/home","timestamp":"2024-03-10T12:02:00Z"}"""
    svc.processEvent("analytics.page.viewed", event1).unsafeRunSync()
    svc.processEvent("analytics.page.viewed", event2).unsafeRunSync()
    svc.processEvent("analytics.page.viewed", event3).unsafeRunSync()
    val metrics = svc.generateUserMetrics("2024-03-10").unsafeRunSync()
    metrics.pageViews      shouldBe 3
    metrics.uniqueSessions shouldBe 2
  }

  test("processEvent product.clicked increments clicks") {
    val (svc, _) = makeService()
    val event = """{"productId":"prod-1","sessionId":"sess-1","eventType":"click","timestamp":"2024-03-10T12:00:00Z"}"""
    svc.processEvent("analytics.product.clicked", event).unsafeRunSync()
    val report = svc.generateDailySales("2024-03-10").unsafeRunSync()
    report.topProducts should have size 1
    report.topProducts.head.productId shouldBe "prod-1"
    report.topProducts.head.clicks    shouldBe 1
    report.topProducts.head.purchases shouldBe 0
  }

  test("processEvent product.clicked with purchase eventType increments revenue") {
    val (svc, _) = makeService()
    val event = """{"productId":"prod-2","sessionId":"sess-1","eventType":"purchase","timestamp":"2024-03-10T12:00:00Z"}"""
    svc.processEvent("analytics.product.clicked", event).unsafeRunSync()
    val report = svc.generateDailySales("2024-03-10").unsafeRunSync()
    report.orderCount    shouldBe 1
    report.totalRevenue  should be > 0.0
  }

  test("generateTopProducts aggregates across date range") {
    val (svc, _) = makeService()
    val click1 = """{"productId":"prod-X","sessionId":"s","eventType":"click","timestamp":"2024-03-01T12:00:00Z"}"""
    val click2 = """{"productId":"prod-X","sessionId":"s","eventType":"click","timestamp":"2024-03-02T12:00:00Z"}"""
    val click3 = """{"productId":"prod-Y","sessionId":"s","eventType":"click","timestamp":"2024-03-01T12:00:00Z"}"""
    svc.processEvent("analytics.product.clicked", click1).unsafeRunSync()
    svc.processEvent("analytics.product.clicked", click2).unsafeRunSync()
    svc.processEvent("analytics.product.clicked", click3).unsafeRunSync()
    val products = svc.generateTopProducts("2024-03-01", "2024-03-31", 10).unsafeRunSync()
    products should have size 2
    products.head.productId shouldBe "prod-X"
    products.head.clicks    shouldBe 2
  }

  test("generateTopProducts respects limit parameter") {
    val (svc, _) = makeService()
    (1 to 5).foreach { i =>
      val ev = s"""{"productId":"prod-$i","sessionId":"s","eventType":"click","timestamp":"2024-04-01T12:00:00Z"}"""
      svc.processEvent("analytics.product.clicked", ev).unsafeRunSync()
    }
    val products = svc.generateTopProducts("2024-04-01", "2024-04-30", 3).unsafeRunSync()
    products should have size 3
  }

  test("processEvent ignores malformed JSON gracefully") {
    val (svc, _) = makeService()
    noException should be thrownBy {
      svc.processEvent("analytics.page.viewed", "not-valid-json{{{").unsafeRunSync()
    }
  }
}
