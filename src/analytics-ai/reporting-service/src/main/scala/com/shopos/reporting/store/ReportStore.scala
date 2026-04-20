package com.shopos.reporting.store

import cats.effect.{IO, Ref}
import com.shopos.reporting.domain._

trait ReportStore {
  def savePageView(date: String, sessionId: String): IO[Unit]
  def saveProductClick(date: String, productId: String, isPurchase: Boolean): IO[Unit]
  def getDailySales(date: String): IO[Option[DailySalesReport]]
  def getUserMetrics(date: String): IO[Option[UserMetrics]]
  def getTopProducts(startDate: String, endDate: String, limit: Int): IO[List[ProductStat]]
}

// Internal state for in-memory store
private final case class StoreState(
    pageViews: Map[String, List[String]],         // date -> list of sessionIds
    productClicks: Map[String, Map[String, Int]], // date -> productId -> clickCount
    productPurchases: Map[String, Map[String, Int]], // date -> productId -> purchaseCount
    revenue: Map[String, Double],                 // date -> totalRevenue
    orderCounts: Map[String, Int]                 // date -> orderCount
)

private object StoreState {
  val empty: StoreState = StoreState(
    pageViews        = Map.empty,
    productClicks    = Map.empty,
    productPurchases = Map.empty,
    revenue          = Map.empty,
    orderCounts      = Map.empty
  )
}

final class InMemoryReportStore(stateRef: Ref[IO, StoreState]) extends ReportStore {

  def savePageView(date: String, sessionId: String): IO[Unit] =
    stateRef.update { state =>
      val existing = state.pageViews.getOrElse(date, List.empty)
      state.copy(pageViews = state.pageViews.updated(date, sessionId :: existing))
    }

  def saveProductClick(date: String, productId: String, isPurchase: Boolean): IO[Unit] =
    stateRef.update { state =>
      val clicks    = state.productClicks.getOrElse(date, Map.empty)
      val purchases = state.productPurchases.getOrElse(date, Map.empty)
      val newClicks = clicks.updated(productId, clicks.getOrElse(productId, 0) + 1)

      if (isPurchase) {
        val newPurchases = purchases.updated(productId, purchases.getOrElse(productId, 0) + 1)
        val rev          = state.revenue.getOrElse(date, 0.0) + 49.99 // simulated revenue per purchase
        val orders       = state.orderCounts.getOrElse(date, 0) + 1
        state.copy(
          productClicks    = state.productClicks.updated(date, newClicks),
          productPurchases = state.productPurchases.updated(date, newPurchases),
          revenue          = state.revenue.updated(date, rev),
          orderCounts      = state.orderCounts.updated(date, orders)
        )
      } else {
        state.copy(productClicks = state.productClicks.updated(date, newClicks))
      }
    }

  def getDailySales(date: String): IO[Option[DailySalesReport]] =
    stateRef.get.map { state =>
      val totalRevenue = state.revenue.getOrElse(date, 0.0)
      val orderCount   = state.orderCounts.getOrElse(date, 0)
      val avgOrderValue =
        if (orderCount == 0) 0.0 else BigDecimal(totalRevenue / orderCount).setScale(2, BigDecimal.RoundingMode.HALF_UP).toDouble

      val clickMap    = state.productClicks.getOrElse(date, Map.empty)
      val purchaseMap = state.productPurchases.getOrElse(date, Map.empty)
      val topProducts = clickMap.keys.toList
        .map { pid =>
          val clicks    = clickMap.getOrElse(pid, 0)
          val purchases = purchaseMap.getOrElse(pid, 0)
          val ctr       = if (clicks == 0) 0.0 else BigDecimal(purchases.toDouble / clicks).setScale(4, BigDecimal.RoundingMode.HALF_UP).toDouble
          ProductStat(pid, clicks, purchases, ctr)
        }
        .sortBy(-_.clicks)
        .take(10)

      Some(DailySalesReport(date, totalRevenue, orderCount, avgOrderValue, topProducts))
    }

  def getUserMetrics(date: String): IO[Option[UserMetrics]] =
    stateRef.get.map { state =>
      val sessions    = state.pageViews.getOrElse(date, List.empty)
      val pageViews   = sessions.size
      val unique      = sessions.distinct.size
      val bounceRate  =
        if (unique == 0) 0.0
        else {
          val singlePageSessions = sessions.groupBy(identity).count { case (_, v) => v.size == 1 }
          BigDecimal(singlePageSessions.toDouble / unique).setScale(4, BigDecimal.RoundingMode.HALF_UP).toDouble
        }
      Some(UserMetrics(date, pageViews, unique, bounceRate))
    }

  def getTopProducts(startDate: String, endDate: String, limit: Int): IO[List[ProductStat]] =
    stateRef.get.map { state =>
      val relevantDates = state.productClicks.keys.filter(d => d >= startDate && d <= endDate).toList

      val aggregated = relevantDates.foldLeft(Map.empty[String, (Int, Int)]) { (acc, date) =>
        val clicks    = state.productClicks.getOrElse(date, Map.empty)
        val purchases = state.productPurchases.getOrElse(date, Map.empty)
        val allProducts = (clicks.keySet ++ purchases.keySet).toList
        allProducts.foldLeft(acc) { (inner, pid) =>
          val (prevClicks, prevPurchases) = inner.getOrElse(pid, (0, 0))
          inner.updated(pid, (prevClicks + clicks.getOrElse(pid, 0), prevPurchases + purchases.getOrElse(pid, 0)))
        }
      }

      aggregated.map { case (pid, (clicks, purchases)) =>
        val ctr = if (clicks == 0) 0.0 else BigDecimal(purchases.toDouble / clicks).setScale(4, BigDecimal.RoundingMode.HALF_UP).toDouble
        ProductStat(pid, clicks, purchases, ctr)
      }.toList.sortBy(-_.clicks).take(limit)
    }
}

object InMemoryReportStore {
  def apply(): IO[InMemoryReportStore] =
    Ref.of[IO, StoreState](StoreState.empty).map(new InMemoryReportStore(_))
}
