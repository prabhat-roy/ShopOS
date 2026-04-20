package com.shopos.reporting.domain

import io.circe.generic.semiauto._
import io.circe.{Decoder, Encoder}

final case class ProductStat(
    productId: String,
    clicks: Int,
    purchases: Int,
    conversionRate: Double
)

object ProductStat {
  implicit val encoder: Encoder[ProductStat] = deriveEncoder
  implicit val decoder: Decoder[ProductStat] = deriveDecoder
}

final case class DailySalesReport(
    date: String,
    totalRevenue: Double,
    orderCount: Int,
    avgOrderValue: Double,
    topProducts: List[ProductStat]
)

object DailySalesReport {
  implicit val encoder: Encoder[DailySalesReport] = deriveEncoder
  implicit val decoder: Decoder[DailySalesReport] = deriveDecoder
}

final case class UserMetrics(
    date: String,
    pageViews: Int,
    uniqueSessions: Int,
    bounceRate: Double
)

object UserMetrics {
  implicit val encoder: Encoder[UserMetrics] = deriveEncoder
  implicit val decoder: Decoder[UserMetrics] = deriveDecoder
}

final case class ReportRequest(
    reportType: String,
    startDate: String,
    endDate: String
)

object ReportRequest {
  implicit val encoder: Encoder[ReportRequest] = deriveEncoder
  implicit val decoder: Decoder[ReportRequest] = deriveDecoder
}

final case class PageViewEvent(
    sessionId: String,
    userId: Option[String],
    pageUrl: String,
    timestamp: String
)

object PageViewEvent {
  implicit val encoder: Encoder[PageViewEvent] = deriveEncoder
  implicit val decoder: Decoder[PageViewEvent] = deriveDecoder
}

final case class ProductClickEvent(
    productId: String,
    sessionId: String,
    userId: Option[String],
    eventType: Option[String],
    timestamp: String
)

object ProductClickEvent {
  implicit val encoder: Encoder[ProductClickEvent] = deriveEncoder
  implicit val decoder: Decoder[ProductClickEvent] = deriveDecoder
}
