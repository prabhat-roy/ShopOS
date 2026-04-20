package com.shopos.reporting.service

import cats.effect.IO
import com.shopos.reporting.domain._
import com.shopos.reporting.store.ReportStore
import io.circe.parser._
import org.typelevel.log4cats.slf4j.Slf4jLogger

trait ReportingService {
  def generateDailySales(date: String): IO[DailySalesReport]
  def generateUserMetrics(date: String): IO[UserMetrics]
  def generateTopProducts(startDate: String, endDate: String, limit: Int): IO[List[ProductStat]]
  def processEvent(topic: String, eventJson: String): IO[Unit]
}

final class ReportingServiceImpl(store: ReportStore) extends ReportingService {
  private val logger = Slf4jLogger.getLogger[IO]

  def generateDailySales(date: String): IO[DailySalesReport] =
    store.getDailySales(date).map(_.getOrElse(DailySalesReport(date, 0.0, 0, 0.0, List.empty)))

  def generateUserMetrics(date: String): IO[UserMetrics] =
    store.getUserMetrics(date).map(_.getOrElse(UserMetrics(date, 0, 0, 0.0)))

  def generateTopProducts(startDate: String, endDate: String, limit: Int): IO[List[ProductStat]] =
    store.getTopProducts(startDate, endDate, limit)

  def processEvent(topic: String, eventJson: String): IO[Unit] = {
    topic match {
      case t if t.contains("page.viewed") =>
        parse(eventJson) match {
          case Right(json) =>
            val cursor    = json.hcursor
            val sessionId = cursor.downField("sessionId").as[String].getOrElse("unknown")
            val timestamp = cursor.downField("timestamp").as[String].getOrElse("")
            val date      = extractDate(timestamp)
            store.savePageView(date, sessionId)
          case Left(err) =>
            logger.warn(s"Failed to parse page.viewed event: $err")
        }

      case t if t.contains("product.clicked") =>
        parse(eventJson) match {
          case Right(json) =>
            val cursor     = json.hcursor
            val productId  = cursor.downField("productId").as[String].getOrElse("unknown")
            val timestamp  = cursor.downField("timestamp").as[String].getOrElse("")
            val eventType  = cursor.downField("eventType").as[String].getOrElse("click")
            val date       = extractDate(timestamp)
            val isPurchase = eventType == "purchase"
            store.saveProductClick(date, productId, isPurchase)
          case Left(err) =>
            logger.warn(s"Failed to parse product.clicked event: $err")
        }

      case unknown =>
        logger.info(s"Ignoring unknown topic: $unknown")
    }
  }

  private def extractDate(timestamp: String): String = {
    if (timestamp.length >= 10) timestamp.take(10)
    else {
      val now = java.time.LocalDate.now().toString
      now
    }
  }
}

object ReportingService {
  def apply(store: ReportStore): ReportingService = new ReportingServiceImpl(store)
}
